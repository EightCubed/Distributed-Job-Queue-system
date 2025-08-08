package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/db"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/logger"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/queue"
	"github.com/EightCubed/Distributed-Job-Queue-system/pkg/models"
	"github.com/alitto/pond/v2"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

const (
	RedisAddr      = "localhost:6379"
	RedisPassword  = ""
	MaxRetries     = 5
	BaseBackoffSec = 5
)

type EmailHandler struct {
	Receiver string `json:"data"`
	Message  string `json:"message"`
}

func (email *EmailHandler) ExecuteJob(log *zap.SugaredLogger, job models.RedisJobType) error {
	err := expectError(config.EMAIL_SUCCESS_CHANCE)
	if err != nil {
		return fmt.Errorf("error during sending email")
	}
	fmt.Printf("Executed Job %v", email)
	return nil
}

func (email *EmailHandler) InitializeHandler(log *zap.SugaredLogger, job models.RedisJobType) {
	email.Receiver = job.Payload.Data
	email.Message = job.Payload.Message
}

type MessageHandler struct {
	Receiver string `json:"data"`
	Message  string `json:"message"`
}

func (msg *MessageHandler) ExecuteJob(log *zap.SugaredLogger, job models.RedisJobType) error {
	err := expectError(config.EMAIL_SUCCESS_CHANCE)
	if err != nil {
		return fmt.Errorf("error during sending message")
	}
	fmt.Printf("Executed Job %v", msg)
	return nil
}

func (msg *MessageHandler) InitializeHandler(log *zap.SugaredLogger, job models.RedisJobType) {
	msg.Receiver = job.Payload.Data
	msg.Message = job.Payload.Message
}

type WebhookHandler struct {
	WebhookURL string `json:"data"`
	Message    string `json:"message"`
}

func (webhook *WebhookHandler) ExecuteJob(log *zap.SugaredLogger, job models.RedisJobType) error {
	err := expectError(config.EMAIL_SUCCESS_CHANCE)
	if err != nil {
		return fmt.Errorf("error during sending webhook")
	}
	fmt.Printf("Executed Job %v", webhook)
	return nil
}

func (webhook *WebhookHandler) InitializeHandler(log *zap.SugaredLogger, job models.RedisJobType) {
	webhook.WebhookURL = job.Payload.Data
	webhook.Message = job.Payload.Message
}

type JobType interface {
	InitializeHandler(*zap.SugaredLogger, models.RedisJobType)
	ExecuteJob(*zap.SugaredLogger, models.RedisJobType) error
}

func main() {
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
	jobQueue := queue.ReturnNewQueue()

	var pollWg sync.WaitGroup
	pollWg.Add(3)

	var handlerWg sync.WaitGroup
	handlerWg.Add(1)

	go PollAndSendJob(ctx, &pollWg, redisClient, models.JOB_PRIORITY_HIGH, config.HIGH_PRIORITY_POLLING_INTERVAL, jobQueue.HighPriorityJobQueue, log)
	go PollAndSendJob(ctx, &pollWg, redisClient, models.JOB_PRIORITY_MEDIUM, config.MEDIUM_PRIORITY_POLLING_INTERVAL, jobQueue.MediumPriorityJobQueue, log)
	go PollAndSendJob(ctx, &pollWg, redisClient, models.JOB_PRIORITY_LOW, config.LOW_PRIORITY_POLLING_INTERVAL, jobQueue.LowPriorityJobQueue, log)

	workerPool := pond.NewPool(50)
	log.Info("Worker Pool started")

	go handleJobs(ctx, &handlerWg, workerPool, jobQueue, log, redisClient)

	go func() {
		pollWg.Wait()
		close(jobQueue.HighPriorityJobQueue)
		close(jobQueue.MediumPriorityJobQueue)
		close(jobQueue.LowPriorityJobQueue)
	}()

	handlerWg.Wait()
	workerPool.StopAndWait()
	log.Info("Worker Pool stopped")
	log.Info("Graceful shutdown complete")
}

func PollAndSendJob(
	ctx context.Context,
	wg *sync.WaitGroup,
	redisClient *redis.Client,
	priority models.JOB_PRIORITY,
	pollingInterval time.Duration,
	jobChan chan<- models.RedisJobType,
	log *zap.SugaredLogger,
) {
	defer wg.Done()
	ticker := time.NewTicker(pollingInterval)
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

			if len(jobs) > 0 {
				log.Infow("Polled jobs", "priority", priority, "count", len(jobs))
			}

			for _, jobStr := range jobs {
				var job models.RedisJobType
				if err := json.Unmarshal([]byte(jobStr), &job); err != nil {
					log.Warnf("Failed to unmarshal job: %v", err)
					continue
				}

				select {
				case jobChan <- job:
					if err := redisClient.ZRem(ctx, key, jobStr).Err(); err != nil {
						log.Warnf("Failed to remove job from Redis: %v", err)
					}
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

var jobRegistry = map[models.JOB_TYPE]JobType{
	models.JOB_TYPE_EMAIL:   &EmailHandler{},
	models.JOB_TYPE_MESSAGE: &MessageHandler{},
	models.JOB_TYPE_WEBHOOK: &WebhookHandler{},
}

func performTask(log *zap.SugaredLogger, job models.RedisJobType, redisClient *redis.Client) {
	jobType := models.JOB_TYPE(job.Type)

	handler, exists := jobRegistry[jobType]
	if !exists {
		log.Errorf("No handler registered for job type: %s", job.Type)
		return
	}

	handler.InitializeHandler(log, job)

	if err := handler.ExecuteJob(log, job); err != nil {
		log.Errorf("Job execution failed for type %s: %v", job.Type, err)

		if job.Retries < MaxRetries {
			job.Retries++
			delay := time.Duration(BaseBackoffSec*(1<<job.Retries)) * time.Second

			executeAt := time.Now().Add(delay).Unix()
			key := fmt.Sprintf("job_%s", job.Priority)

			jobBytes, _ := json.Marshal(job)
			if err := redisClient.ZAdd(context.Background(), key, &redis.Z{
				Score:  float64(executeAt),
				Member: string(jobBytes),
			}).Err(); err != nil {
				log.Errorf("Failed to requeue job: %v", err)
			} else {
				log.Infof("Requeued job %s for retry #%d after %v", job.Type, job.Retries, delay)
			}
		} else {
			log.Warnf("Max retries reached for job %s. Dropping job.", job.Type)
		}
	} else {
		log.Infof("Job executed successfully: %s", job.Type)
	}
}

func handleJobs(
	ctx context.Context,
	wg *sync.WaitGroup,
	pool pond.Pool,
	jobQueue *queue.JobQueueType,
	log *zap.SugaredLogger,
	redisClient *redis.Client,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			log.Info("Stopping job handler")
			return
		case job, ok := <-jobQueue.HighPriorityJobQueue:
			if !ok {
				jobQueue.HighPriorityJobQueue = nil
				continue
			}
			pool.Submit(func() { performTask(log, job, redisClient) })
		case job, ok := <-jobQueue.MediumPriorityJobQueue:
			if !ok {
				jobQueue.MediumPriorityJobQueue = nil
				continue
			}
			pool.Submit(func() { performTask(log, job, redisClient) })
		case job, ok := <-jobQueue.LowPriorityJobQueue:
			if !ok {
				jobQueue.LowPriorityJobQueue = nil
				continue
			}
			pool.Submit(func() { performTask(log, job, redisClient) })
		}

		if jobQueue.HighPriorityJobQueue == nil &&
			jobQueue.MediumPriorityJobQueue == nil &&
			jobQueue.LowPriorityJobQueue == nil {
			log.Info("All job queues closed, handler exiting")
			return
		}
	}
}

func expectError(chance int) error {
	if chance < rand.Intn(100) {
		return fmt.Errorf("Error")
	}
	return nil
}
