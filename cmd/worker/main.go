package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/db"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/logger"
	"github.com/EightCubed/Distributed-Job-Queue-system/pkg/models"
	"github.com/alitto/pond/v2"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	RedisAddr     = "localhost:6379"
	RedisPassword = ""
)

func main() {
	// create common workers that polls redis for a certain time and adds tasks to one of three channels
	lgr, err := logger.SetupLogger()
	if err != nil {
		panic(err)
	}
	log := logger.FetchSugaredLogger(lgr)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info("Shutdown signal received")
		cancel()
	}()

	redisClient := db.NewRedisClient(RedisAddr, RedisPassword)

	highPriorityJobChannel := make(chan models.RedisJobType, config.BATCH_SIZE)
	mediumPriorityJobChannel := make(chan models.RedisJobType, config.BATCH_SIZE)
	lowPriorityJobChannel := make(chan models.RedisJobType, config.BATCH_SIZE)

	var pollWg sync.WaitGroup
	pollWg.Add(3)

	var handlerWg sync.WaitGroup
	handlerWg.Add(3)

	go PollAndSendJob(ctx, &pollWg, redisClient, models.JOB_PRIORITY_HIGH, config.HIGH_PRIORITY_POLLING_INTERVAL, highPriorityJobChannel, log)
	go PollAndSendJob(ctx, &pollWg, redisClient, models.JOB_PRIORITY_MEDIUM, config.MEDIUM_PRIORITY_POLLING_INTERVAL, mediumPriorityJobChannel, log)
	go PollAndSendJob(ctx, &pollWg, redisClient, models.JOB_PRIORITY_LOW, config.LOW_PRIORITY_POLLING_INTERVAL, lowPriorityJobChannel, log)

	highPriorityPool := pond.NewPool(100)
	mediumPriorityPool := pond.NewPool(500)
	lowPriorityPool := pond.NewPool(1000)
	log.Infof("%s Priority Pool started", models.JOB_PRIORITY_HIGH)
	log.Infof("%s Priority Pool started", models.JOB_PRIORITY_MEDIUM)
	log.Infof("%s Priority Pool started", models.JOB_PRIORITY_LOW)

	go handleJobs(&handlerWg, highPriorityPool, highPriorityJobChannel, log)
	go handleJobs(&handlerWg, mediumPriorityPool, mediumPriorityJobChannel, log)
	go handleJobs(&handlerWg, lowPriorityPool, lowPriorityJobChannel, log)

	go func() {
		pollWg.Wait()
		close(highPriorityJobChannel)
		close(mediumPriorityJobChannel)
		close(lowPriorityJobChannel)
	}()

	handlerWg.Wait()
	highPriorityPool.StopAndWait()
	log.Infof("%s Priority Pool stopped", models.JOB_PRIORITY_HIGH)
	mediumPriorityPool.StopAndWait()
	log.Infof("%s Priority Pool stopped", models.JOB_PRIORITY_MEDIUM)
	lowPriorityPool.StopAndWait()
	log.Infof("%s Priority Pool stopped", models.JOB_PRIORITY_LOW)

	log.Info("Graceful shutdown complete")
}

func PollAndSendJob(ctx context.Context, wg *sync.WaitGroup, redisClient *redis.Client, priority models.JOB_PRIORITY, polling_interval time.Duration, jobChan chan<- models.RedisJobType, log *zap.SugaredLogger) {
	defer wg.Done()
	ticker := time.NewTicker(polling_interval)
	defer ticker.Stop()

	key := fmt.Sprintf("job_%s", priority)

	for {
		select {
		case <-ctx.Done():
			log.Infof("Polling stopped for %s", priority)
			return

		case <-ticker.C:
			now := float64(time.Now().Unix())

			jobs, err := redisClient.ZRangeByScore(ctx, key, &redis.ZRangeBy{
				Min:   "0",
				Max:   fmt.Sprintf("%.0f", now),
				Count: config.BATCH_SIZE,
			}).Result()
			if err != nil {
				log.Warnf("Redis poll error [%s]: %v", priority, err)
				continue
			}

			log.Infow("Polled jobs",
				"priority", priority,
				"count", len(jobs),
			)

			for _, jobStr := range jobs {
				var job models.RedisJobType
				if err := json.Unmarshal([]byte(jobStr), &job); err != nil {
					log.Warnf("Failed to unmarshal job: %v", err)
					continue
				}

				jobChan <- job
				if err := redisClient.ZRem(ctx, key, jobStr).Err(); err != nil {
					log.Warnf("Failed to remove job from Redis: %v", err)
				}
			}
		}
	}
}

func performTask(job models.RedisJobType, log *zap.SugaredLogger) {
	log.Infow("Job Executed",
		"type", job.Type,
		"data", job.Payload.Data,
		"message", job.Payload.Message,
	)
}

func handleJobs(
	wg *sync.WaitGroup,
	pool pond.Pool,
	jobChan <-chan models.RedisJobType,
	log *zap.SugaredLogger,
) {
	defer wg.Done()

	for job := range jobChan {
		j := job // avoid closure bug
		pool.Submit(func() {
			performTask(j, log)
		})
	}
}
