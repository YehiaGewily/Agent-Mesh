package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/YehiaGewily/Agent-Mesh/internal/config"
	"github.com/YehiaGewily/Agent-Mesh/internal/models"
	"github.com/YehiaGewily/Agent-Mesh/pkg/broker"
	"github.com/YehiaGewily/Agent-Mesh/pkg/database"
	"github.com/YehiaGewily/Agent-Mesh/pkg/notifications"
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
	// Subscriptions...
	go func() {
		pubsub := redisBroker.SubscribeSystemHealth(context.Background())
		defer pubsub.Close()
		ch := pubsub.Channel()
		for msg := range ch {
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

	// CHECK FOR SIMULATION MODE
	if os.Getenv("ENABLE_SIMULATOR") == "true" {
		log.Println("⚠️  SIMULATION MODE ENABLED")
		go p.startSimulation()
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

// startSimulation runs a background loop to generate random tasks
func (p *Producer) startSimulation() {
	// Seed random generator
	// Note: In Go 1.20+ the global seed is auto-initialized, but for older versions:
	// rand.Seed(time.Now().UnixNano())

	agentTypes := []string{models.AgentTypeArchitect, models.AgentTypeDeveloper, models.AgentTypeQA}

	for {
		// Random Jitter: 3s to 7s
		jitter := time.Duration(rand.Intn(4000)+3000) * time.Millisecond
		time.Sleep(jitter)

		// Create Random Task
		agentType := agentTypes[rand.Intn(len(agentTypes))]
		priority := rand.Intn(5) + 1 // 1-5

		task := &models.Task{
			ID:        uuid.New().String(),
			Status:    models.TaskStatusPending,
			Priority:  priority,
			AgentType: agentType,
			Payload: map[string]interface{}{
				"source": "simulator",
				"ts":     time.Now().Unix(),
				"note":   "Automated drill",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := p.CreateTask(context.Background(), task); err != nil {
			log.Printf("[SIMULATOR] Failed to create task: %v", err)
			continue
		}

		log.Printf("[SIMULATOR] Generated Task %s (%s, P%d)", task.ID, task.AgentType, task.Priority)
	}
}

// CreateTask handles persistence, enqueueing, and broadcasting
func (p *Producer) CreateTask(ctx context.Context, task *models.Task) error {
	// 1. Persist to DB
	if err := p.DB.StoreTask(ctx, task); err != nil {
		return fmt.Errorf("db store failed: %w", err)
	}

	// 2. Enqueue to Redis
	if err := p.Broker.Enqueue(ctx, task.ID, task.Priority); err != nil {
		return fmt.Errorf("redis enqueue failed: %w", err)
	}

	// 3. Broadcast Event
	if err := p.Broker.PublishTaskEvent(ctx, task); err != nil {
		log.Printf("Failed to broadcast task event: %v", err)
		// Non-critical
	}

	return nil
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
	if req.AgentType != models.AgentTypeArchitect &&
		req.AgentType != models.AgentTypeDeveloper &&
		req.AgentType != models.AgentTypeQA {
		http.Error(w, fmt.Sprintf("Invalid agent_type. Must be one of: %s, %s, %s",
			models.AgentTypeArchitect, models.AgentTypeDeveloper, models.AgentTypeQA), http.StatusBadRequest)
		return
	}

	// Create Task Object
	task := &models.Task{
		ID:        uuid.New().String(),
		Status:    models.TaskStatusPending,
		Priority:  req.Priority,
		AgentType: req.AgentType,
		Payload:   req.Payload,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Use Shared Logic
	if err := p.CreateTask(r.Context(), task); err != nil {
		log.Printf("CreateTask failed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Respond
	w.WriteHeader(http.StatusAccepted)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TaskResponse{
		ID:     task.ID,
		Status: string(task.Status),
	})

	log.Printf("Task %s accepted for agent %s (Priority %d)", task.ID, task.AgentType, task.Priority)
}
