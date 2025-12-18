package worker

import (
	"context"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"github.com/YehiaGewily/agentmesh/internal/models"
	"github.com/YehiaGewily/agentmesh/pkg/broker"
	"github.com/YehiaGewily/agentmesh/pkg/database"
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

	if task.AgentType == models.AgentTypeMagnus {
		log.Printf("[Worker %d] Starting Magnus Strategist Agent...", workerID)
	}

	// 2. Simulate AI Processing
	err = w.simulateAIWork(task)

	if err == nil {
		// Success
		if err := w.DB.UpdateTaskStatus(ctx, task.ID, models.TaskStatusCompleted); err != nil {
			log.Printf("[Worker %d] Failed to mark task %s completed: %v", workerID, task.ID, err)
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
