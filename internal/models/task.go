package models

import "time"

// Agent Types (The Rankup Squad)
const (
	AgentTypeArchitect = "ARCHITECT"
	AgentTypeDeveloper = "DEVELOPER"
	AgentTypeQA        = "QA_ENGINEER"
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
type SystemHealth struct {
	ReqType   string  `json:"type"` // "HEALTH_METRIC"
	WorkerID  int     `json:"worker_id"`
	CPUUsage  float64 `json:"cpu_usage"`
	RAMUsage  float64 `json:"ram_usage"` // Used Percent of Soft Limit (512MB)
	RAMUsedMB float64 `json:"ram_used_mb"`
	Timestamp string  `json:"timestamp"`
}
