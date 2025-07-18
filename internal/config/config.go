package config

import (
	"context"

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
