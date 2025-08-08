package config

import (
	"context"
	"time"

	"go.uber.org/zap"
)

type Config struct {
	ServerPort    string
	RedisAddr     string
	RedisPassword string
	PostgresDSN   string
}

type ctxKey string

const LoggerKey ctxKey = "logger"

func LoggerFromContext(ctx context.Context) *zap.Logger {
	val := ctx.Value(LoggerKey)
	if logger, ok := val.(*zap.Logger); ok {
		return logger
	}
	return zap.NewNop()
}

const (
	HIGH_PRIORITY_POLLING_INTERVAL   time.Duration = 3 * time.Second
	MEDIUM_PRIORITY_POLLING_INTERVAL time.Duration = 30 * time.Second
	LOW_PRIORITY_POLLING_INTERVAL    time.Duration = 90 * time.Second
)

const BATCH_SIZE = 10000

const (
	EMAIL_SUCCESS_CHANCE   = 60
	MESSAGE_SUCCESS_CHANCE = 80
	WEBHOOK_SUCCESS_CHANCE = 95
)
