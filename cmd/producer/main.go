package main

import (
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
)

type Producer struct {
	Broker *broker.RedisBroker
	DB     *database.DB
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

	p := &Producer{
		Broker: redisBroker,
		DB:     db,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/v1/tasks", p.handleCreateTask)

	log.Println("Producer API listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
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

	// 3. Respond
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:     task.ID,
		Status: string(task.Status),
	})

	log.Printf("Task %s accepted for agent %s (Priority %d)", task.ID, task.AgentType, task.Priority)
}
