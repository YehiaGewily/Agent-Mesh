package main

import (
	"fmt"
	"log"

	"github.com/YehiaGewily/agentmesh/internal/config"
)

func main() {
	cfg := config.Load()
	fmt.Printf("Starting Producer Service...\n")
	fmt.Printf("Connected to Redis at: %s\n", cfg.RedisAddr)

	// Todo: Implement API
	log.Println("Producer running (press Ctrl+C to stop)")
	select {}
}
