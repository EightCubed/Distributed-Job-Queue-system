package models

import "time"

type Job struct {
	ID          int       `json:"id"`
	Type        string    `json:"type"`
	Data        string    `json:"data"`
	Message     string    `json:"message"`
	Priority    string    `json:"priority"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ExecutionAt time.Time `json:"execution_at"`
}

type JOB_STATUS string

const (
	JOB_STATUS_NONE      JOB_STATUS = "none"
	JOB_STATUS_QUEUED    JOB_STATUS = "queued"
	JOB_STATUS_FAILED    JOB_STATUS = "failed"
	JOB_STATUS_PROGRESS  JOB_STATUS = "progress"
	JOB_STATUS_COMPLETED JOB_STATUS = "completed"
)

type JOB_PRIORITY string

const (
	JOB_PRIORITY_HIGH   JOB_PRIORITY = "HIGH"
	JOB_PRIORITY_MEDIUM JOB_PRIORITY = "MEDIUM"
	JOB_PRIORITY_LOW    JOB_PRIORITY = "LOW"
)

type JobBody struct {
	Type     string       `json:"type"`
	Payload  PayloadType  `json:"payload"`
	Priority JOB_PRIORITY `json:"priority"`
	Delay    int          `json:"delay"`
}

type RedisJobType struct {
	Type        string      `json:"type"`
	Payload     PayloadType `json:"payload"`
	ExecutionAt time.Time   `json:"execution_at"`
}

type PayloadType struct {
	Data    string `json:"data"`
	Message string `json:"message"`
}
