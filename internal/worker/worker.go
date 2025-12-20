package worker

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/process"

	"github.com/YehiaGewily/Agent-Mesh/internal/models"
	"github.com/YehiaGewily/Agent-Mesh/pkg/broker"
	"github.com/YehiaGewily/Agent-Mesh/pkg/database"
)

type Worker struct {
	Broker *broker.RedisBroker
	DB     *database.DB
}

func NewWorker(b *broker.RedisBroker, db *database.DB) *Worker {
	return &Worker{
		Broker: b,
		DB:     db,
	}
}

func (w *Worker) Start(ctx context.Context, concurrency int) {
	var wg sync.WaitGroup
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			w.loop(ctx, workerID)
		}(i)
	}
	wg.Wait()
}

func (w *Worker) StartHealthMonitor(ctx context.Context, workerID int) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	log.Printf("[HealthMonitor %d] Started", workerID)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// CPU Usage
			cpuPercent, err := cpu.Percent(0, false)
			if err != nil {
				log.Printf("Error getting CPU: %v", err)
				continue
			}

			// Process RAM Usage
			proc, err := process.NewProcess(int32(os.Getpid()))
			var ramMb float64
			var ramPercent float64

			if err == nil {
				memInfo, err := proc.MemoryInfo()
				if err == nil {
					ramMb = float64(memInfo.RSS) / 1024 / 1024
					// 512MB Soft Limit for bar calculation
					ramPercent = (ramMb / 512.0) * 100
				} else {
					log.Printf("Error getting process memory: %v", err)
				}
			} else {
				log.Printf("Error getting process: %v", err)
			}

			health := &models.SystemHealth{
				ReqType:   "HEALTH_METRIC",
				WorkerID:  workerID,
				CPUUsage:  cpuPercent[0],
				RAMUsage:  ramPercent,
				RAMUsedMB: ramMb,
				Timestamp: time.Now().Format(time.RFC3339),
			}

			if err := w.Broker.PublishSystemHealth(ctx, health); err != nil {
				log.Printf("Error publishing health: %v", err)
			}
		}
	}
}

func (w *Worker) loop(ctx context.Context, workerID int) {
	log.Printf("[Worker %d] Started", workerID)
	for {
		select {
		case <-ctx.Done():
			log.Printf("[Worker %d] Stopping", workerID)
			return
		default:
			// Fetch task (blocking)
			taskID, err := w.Broker.FetchTask(ctx)
			if err != nil {
				// Don't spam logs if it's just a timeout or context cancel
				if ctx.Err() != nil {
					return
				}
				log.Printf("[Worker %d] Error fetching task: %v", workerID, err)
				time.Sleep(1 * time.Second)
				continue
			}

			w.processTask(ctx, workerID, taskID)
		}
	}
}

func (w *Worker) processTask(ctx context.Context, workerID int, taskID string) {
	// 1. Get Task Details
	task, err := w.DB.GetTask(ctx, taskID)
	if err != nil {
		log.Printf("[Worker %d] Failed to get task %s details: %v", workerID, taskID, err)
		return
	}

	log.Printf("[Worker %d] Processing task %s for %s (Priority: %d)", workerID, task.ID, task.AgentType, task.Priority)

	switch task.AgentType {
	case models.AgentTypeArchitect:
		log.Printf("[Worker %d] Starting System Architecture Analysis...", workerID)
	case models.AgentTypeDeveloper:
		log.Printf("[Worker %d] Writing Code Implementation...", workerID)
	case models.AgentTypeQA:
		log.Printf("[Worker %d] Running Test Suite...", workerID)
	}

	// 2. Simulate AI Processing
	err = w.simulateAIWork(task)

	if err == nil {
		// Success
		now := time.Now()
		if err := w.DB.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted); err != nil {
			log.Printf("[Worker %d] Failed to mark task %s completed: %v", workerID, task.ID, err)
		}

		// Update struct for broadcast
		task.Status = models.TaskStatusCompleted
		task.UpdatedAt = now

		// Broadcast Completion Event
		if err := w.Broker.PublishTaskEvent(ctx, task); err != nil {
			log.Printf("[Worker %d] Failed to broadcast completion for %s: %v", workerID, task.ID, err)
		}

		log.Printf("[Worker %d] Task %s completed successfully", workerID, task.ID)
		return
	}

	// Failure Handling
	log.Printf("[Worker %d] Task %s failed: %v", workerID, task.ID, err)

	newRetryCount, err := w.DB.IncrementRetryCount(ctx, task.ID)
	if err != nil {
		log.Printf("[Worker %d] Failed to increment retry count for %s: %v", workerID, task.ID, err)
		// Try to at least fail it locally or proceed to backoff if possible?
		// If DB is down, we are in trouble.
	}

	if newRetryCount > 5 {
		// DLQ
		log.Printf("[Worker %d] Task %s exceeded max retries (%d). Moving to DLQ.", workerID, task.ID, newRetryCount)
		if err := w.Broker.AddToDLQ(ctx, task.ID); err != nil {
			log.Printf("[Worker %d] Failed to add %s to DLQ: %v", workerID, task.ID, err)
		}
		if err := w.DB.UpdateTaskStatus(ctx, taskID, models.TaskPermanentFail); err != nil {
			log.Printf("[Worker %d] Failed to mark %s as PERMANENT_FAILURE: %v", workerID, task.ID, err)
		}
	} else {
		// Exponential Backoff
		backoffDuration := time.Duration(math.Pow(2, float64(newRetryCount))) * time.Second
		log.Printf("[Worker %d] Re-queueing task %s in %v", workerID, task.ID, backoffDuration)

		time.Sleep(backoffDuration)

		// Re-enqueue (keep original priority)
		if err := w.Broker.Enqueue(ctx, task.ID, task.Priority); err != nil {
			log.Printf("[Worker %d] Failed to re-enqueue task %s: %v", workerID, task.ID, err)
		}
	}
}

func (w *Worker) simulateAIWork(task *models.Task) error {
	// Simulate AI Agent call
	time.Sleep(2 * time.Second)

	// Simulate random failure for demonstration?
	// The prompt implies "If the 'agent' fails", so I should probably make it fail sometimes?
	// or just assumes external failure.
	// "If the 'agent' fails, the worker should wait..."
	// I will just return nil (success) by default, or maybe check payload for a "fail" flag for testing.

	if val, ok := task.Payload["simulate_fail"]; ok {
		if fail, ok := val.(bool); ok && fail {
			return fmt.Errorf("simulated AI agent error")
		}
	}

	return nil
}
