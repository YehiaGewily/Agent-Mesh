package models

import "time"

// Agent Types (The Rankup Squad)
const (
	AgentTypeMagnus = "MAGNUS_STRATEGIST"
	AgentTypeCedric = "CEDRIC_WRITER"
	AgentTypeLyra   = "LYRA_AUDITOR"
)

type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskPermanentFail   TaskStatus = "PERMANENT_FAILURE"
)

type Task struct {
	ID         string                 `json:"id"`
	Status     TaskStatus             `json:"status"`
	Priority   int                    `json:"priority"`
	AgentType  string                 `json:"agent_type"`
	Payload    map[string]interface{} `json:"payload"`
	RetryCount int                    `json:"retry_count"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
}
