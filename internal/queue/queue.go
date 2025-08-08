package queue

import (
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/EightCubed/Distributed-Job-Queue-system/pkg/models"
)

type JobQueueType struct {
	HighPriorityJobQueue   chan models.RedisJobType
	MediumPriorityJobQueue chan models.RedisJobType
	LowPriorityJobQueue    chan models.RedisJobType
}

func ReturnNewQueue() *JobQueueType {
	highPriorityJobQueue := make(chan models.RedisJobType, config.BATCH_SIZE)
	mediumPriorityJobQueue := make(chan models.RedisJobType, config.BATCH_SIZE)
	lowPriorityJobQueue := make(chan models.RedisJobType, config.BATCH_SIZE)
	return &JobQueueType{
		HighPriorityJobQueue:   highPriorityJobQueue,
		MediumPriorityJobQueue: mediumPriorityJobQueue,
		LowPriorityJobQueue:    lowPriorityJobQueue,
	}
}
