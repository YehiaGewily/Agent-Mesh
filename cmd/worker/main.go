package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/YehiaGewily/Agent-Mesh/internal/config"
	"github.com/YehiaGewily/Agent-Mesh/internal/worker"
	"github.com/YehiaGewily/Agent-Mesh/pkg/broker"
	"github.com/YehiaGewily/Agent-Mesh/pkg/database"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Starting Worker Service...\n")

	// 1. Initialize DB
	db, err := database.NewConnection(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()
	fmt.Printf("Connected to DB\n")

	// 2. Initialize Broker
	redisBroker := broker.NewBroker(cfg.RedisAddr, db)
	fmt.Printf("Connected to Redis at: %s\n", cfg.RedisAddr)

	// 3. Initialize Worker
	w := worker.NewWorker(redisBroker, db)

	// 4. Start Worker Loop with Graceful Custom
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle SIGINT/SIGTERM
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down worker...")
		cancel()
	}()

	// Start Health Monitor
	go w.StartHealthMonitor(ctx, 1) // Using ID 1 for single node monitoring for now

	// Start with 5 concurrent workers
	concurrency := 5
	log.Printf("Starting %d concurrent workers...", concurrency)
	w.Start(ctx, concurrency)

	log.Println("Worker Stopped")
}
