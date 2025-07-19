package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/EightCubed/Distributed-Job-Queue-system/pkg/models"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
)

func (handler *ApiHandler) SubmitJob(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := config.LoggerFromContext(ctx)
	sugar := logger.Sugar()

	sugar.Info("Received job submission")

	var body models.JobBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		sugar.Warnf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if body.Payload.Data == "" || body.Payload.Message == "" {
		sugar.Warnf("Data or Message is empty")
		http.Error(w, "Data or Message is empty", http.StatusBadRequest)
		return
	}

	query := `
		INSERT INTO jobs (type, data, message, priority, delay_seconds, created_at, execution_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id
	`

	createdAt := time.Now().UTC()
	executionAt := createdAt.Add(time.Duration(body.Delay) * time.Second)

	var jobID int
	err := handler.PostgresPool.QueryRow(ctx, query,
		body.Type,
		body.Payload.Data,
		body.Payload.Message,
		body.Priority,
		body.Delay,
		createdAt,
		executionAt,
	).Scan(&jobID)

	if err != nil {
		logger.Error("Database insert failed", zap.Any("request_id", ctx.Value("request_id")))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	job := models.RedisJobType{
		Type:        body.Type,
		Payload:     body.Payload,
		ExecutionAt: executionAt,
	}

	jobJSON, err := json.Marshal(job)
	if err != nil {
		logger.Error("Failed to marshal job for Redis", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	redisKey := fmt.Sprintf("job_%s", body.Priority)
	score := float64(executionAt.Unix())

	redisCmd := handler.RedisClient.ZAdd(ctx, redisKey, &redis.Z{
		Score:  score,
		Member: jobJSON,
	})

	if err := redisCmd.Err(); err != nil {
		logger.Error("Failed to push job to Redis", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Job submitted successfully"))
	logger.Info("Job inserted into database")

}
