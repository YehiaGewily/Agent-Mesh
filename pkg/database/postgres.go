package database

import (
	"context"
	"fmt"
	"time"

	"github.com/YehiaGewily/Agent-Mesh/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func NewConnection(dsn string) (*DB, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse dsn: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) StoreTask(ctx context.Context, task *models.Task) error {
	query := `
		INSERT INTO tasks (id, status, priority, agent_type, payload, retry_count, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := db.Pool.Exec(ctx, query,
		task.ID,
		task.Status,
		task.Priority,
		task.AgentType,
		task.Payload,
		task.RetryCount,
		task.CreatedAt,
		task.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert task: %w", err)
	}
	return nil
}

func (db *DB) UpdateTaskStatus(ctx context.Context, taskID string, status models.TaskStatus) error {
	query := `
		UPDATE tasks 
		SET status = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := db.Pool.Exec(ctx, query, status, time.Now(), taskID)
	if err != nil {
		return fmt.Errorf("failed to update task status: %w", err)
	}
	return nil
}

func (db *DB) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	query := `
		SELECT id, status, priority, agent_type, payload, retry_count, created_at, updated_at
		FROM tasks
		WHERE id = $1
	`
	row := db.Pool.QueryRow(ctx, query, taskID)

	var task models.Task
	err := row.Scan(
		&task.ID,
		&task.Status,
		&task.Priority,
		&task.AgentType,
		&task.Payload,
		&task.RetryCount,
		&task.CreatedAt,
		&task.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan task: %w", err)
	}
	return &task, nil
}

func (db *DB) IncrementRetryCount(ctx context.Context, taskID string) (int, error) {
	query := `
		UPDATE tasks
		SET retry_count = retry_count + 1, updated_at = $1
		WHERE id = $2
		RETURNING retry_count
	`
	var newCount int
	err := db.Pool.QueryRow(ctx, query, time.Now(), taskID).Scan(&newCount)
	if err != nil {
		return 0, fmt.Errorf("failed to increment retry count: %w", err)
	}
	return newCount, nil
}
