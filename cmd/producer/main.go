package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/YehiaGewily/agentmesh/internal/config"
	"github.com/YehiaGewily/agentmesh/internal/models"
	"github.com/YehiaGewily/agentmesh/pkg/broker"
	"github.com/YehiaGewily/agentmesh/pkg/database"
	"github.com/YehiaGewily/agentmesh/pkg/notifications"
)

type Producer struct {
	Broker *broker.RedisBroker
	DB     *database.DB
	Hub    *notifications.Hub
}

type TaskRequest struct {
	AgentType string                 `json:"agent_type"`
	Priority  int                    `json:"priority"`
	Payload   map[string]interface{} `json:"payload"`
}

type TaskResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func main() {
	cfg := config.Load()
	fmt.Printf("Starting Producer Service...\n")

	// Initialize DB
	db, err := database.NewConnection(cfg.DBDSN)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// Initialize Broker
	redisBroker := broker.NewBroker(cfg.RedisAddr, db)
	fmt.Printf("Connected to Redis at: %s\n", cfg.RedisAddr)

	// Initialize Notification Hub
	hub := notifications.NewHub()
	go hub.Run()

	// Subscribe to Redis Updates and broadcast to Hub
	// Subscribe to Redis Updates and broadcast to Hub
	go func() {
		pubsub := redisBroker.SubscribeTaskUpdates(context.Background())
		defer pubsub.Close()
		ch := pubsub.Channel()
		for msg := range ch {
			hub.Broadcast([]byte(msg.Payload))
		}
	}()

	// Subscribe to System Health and broadcast to Hub
	go func() {
		pubsub := redisBroker.SubscribeSystemHealth(context.Background())
		defer pubsub.Close()
		ch := pubsub.Channel()
		for msg := range ch {
			// Wrap in standardized envelope
			wrapper := fmt.Sprintf(`{"type":"HEALTH_UPDATE","data":%s,"timestamp":"%s"}`,
				msg.Payload,
				time.Now().Format(time.RFC3339))

			hub.Broadcast([]byte(wrapper))
		}
	}()

	p := &Producer{
		Broker: redisBroker,
		DB:     db,
		Hub:    hub,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks", p.handleCreateTask)
	mux.HandleFunc("/v1/ws", func(w http.ResponseWriter, r *http.Request) {
		p.Hub.ServeWs(w, r)
	})

	log.Println("Producer API listening on :8081 (WS at /v1/ws)")
	if err := http.ListenAndServe(":8081", mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func (p *Producer) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req TaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Validate Agent Type
	if req.AgentType != models.AgentTypeMagnus &&
		req.AgentType != models.AgentTypeCedric &&
		req.AgentType != models.AgentTypeLyra {
		http.Error(w, fmt.Sprintf("Invalid agent_type. Must be one of: %s, %s, %s",
			models.AgentTypeMagnus, models.AgentTypeCedric, models.AgentTypeLyra), http.StatusBadRequest)
		return
	}

	// Create Task
	task := &models.Task{
		ID:        uuid.New().String(),
		Status:    models.TaskStatusPending,
		Priority:  req.Priority,
		AgentType: req.AgentType,
		Payload:   req.Payload,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// 1. Persist to DB
	if err := p.DB.StoreTask(r.Context(), task); err != nil {
		log.Printf("Failed to store task: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// 2. Enqueue to Redis
	if err := p.Broker.Enqueue(r.Context(), task.ID, task.Priority); err != nil {
		log.Printf("Failed to enqueue task: %v", err)
		http.Error(w, "Failed to enqueue task", http.StatusInternalServerError)
		return
	}

	// 3. Broadcast Event
	if err := p.Broker.PublishTaskEvent(r.Context(), task); err != nil {
		log.Printf("Failed to broadcast task event: %v", err)
		// Non-critical, continue
	}

	// 4. Respond
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:     task.ID,
		Status: string(task.Status),
	})

	log.Printf("Task %s accepted for agent %s (Priority %d)", task.ID, task.AgentType, task.Priority)
}
