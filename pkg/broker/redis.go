package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/YehiaGewily/agentmesh/internal/models"
	"github.com/YehiaGewily/agentmesh/pkg/database"
	"github.com/redis/go-redis/v9"
)

const (
	QueueHigh   = "agent_high"
	QueueMedium = "agent_medium"
	QueueLow    = "agent_low"
)

type RedisBroker struct {
	Client *redis.Client
	DB     *database.DB
}

func NewBroker(addr string, db *database.DB) *RedisBroker {
	rdb := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisBroker{
		Client: rdb,
		DB:     db,
	}
}

func (b *RedisBroker) Enqueue(ctx context.Context, taskID string, priority int) error {
	queue := QueueLow
	if priority >= 3 {
		queue = QueueHigh
	} else if priority == 2 {
		queue = QueueMedium
	}

	err := b.Client.LPush(ctx, queue, taskID).Err()
	if err != nil {
		return fmt.Errorf("failed to enqueue task to %s: %w", queue, err)
	}
	return nil
}

// FetchTask blocks until a task is available in the specified queues,
// then immediately updates its status to 'running' in Postgres (Claim pattern).
func (b *RedisBroker) FetchTask(ctx context.Context, queues ...string) (string, error) {
	// Default priority order if no queues provided
	if len(queues) == 0 {
		queues = []string{QueueHigh, QueueMedium, QueueLow}
	}

	// BLPop or BRPop. User asked for BRPOP.
	// Redis BRPOP returns [listName, value]
	result, err := b.Client.BRPop(ctx, 0*time.Second, queues...).Result()
	if err != nil {
		return "", fmt.Errorf("failed to fetch task: %w", err)
	}

	if len(result) < 2 {
		return "", fmt.Errorf("invalid BRPOP result")
	}

	taskID := result[1]

	// Claim Pattern: Update status to running immediately
	err = b.DB.UpdateTaskStatus(ctx, taskID, models.TaskStatusRunning)
	if err != nil {
		return "", fmt.Errorf("failed to claim task %s: %w", taskID, err)
	}

	// Notify Real-Time (Running)
	b.PublishTaskUpdate(ctx, taskID, string(models.TaskStatusRunning))

	return taskID, nil
}

func (b *RedisBroker) AddToDLQ(ctx context.Context, taskID string) error {
	err := b.Client.RPush(ctx, "agent_dead_letter", taskID).Err()
	if err != nil {
		return fmt.Errorf("failed to add to DLQ: %w", err)
	}
	return nil
}

func (b *RedisBroker) PublishTaskUpdate(ctx context.Context, taskID, status string) error {
	msg := fmt.Sprintf(`{"task_id":"%s","status":"%s"}`, taskID, status)
	err := b.Client.Publish(ctx, "task_updates", msg).Err()
	if err != nil {
		return fmt.Errorf("failed to publish update: %w", err)
	}
	return nil
}

func (b *RedisBroker) PublishTaskEvent(ctx context.Context, task *models.Task) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	err = b.Client.Publish(ctx, "task_updates", data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish task event: %w", err)
	}
	return nil
}

func (b *RedisBroker) PublishSystemHealth(ctx context.Context, health *models.SystemHealth) error {
	data, err := json.Marshal(health)
	if err != nil {
		return fmt.Errorf("failed to marshal health metric: %w", err)
	}

	err = b.Client.Publish(ctx, "system_health", data).Err()
	if err != nil {
		return fmt.Errorf("failed to publish health metric: %w", err)
	}
	return nil
}

func (b *RedisBroker) SubscribeSystemHealth(ctx context.Context) *redis.PubSub {
	return b.Client.Subscribe(ctx, "system_health")
}

func (b *RedisBroker) SubscribeTaskUpdates(ctx context.Context) *redis.PubSub {
	return b.Client.Subscribe(ctx, "task_updates")
}
