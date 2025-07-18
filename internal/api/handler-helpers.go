package api

import (
	"github.com/go-redis/redis/v8"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type ApiHandler struct {
	Logger       *zap.Logger
	PostgresPool *pgxpool.Pool
	RedisClient  *redis.Client
}

type AppContext struct {
	Logger       *zap.Logger
	PostgresPool *pgxpool.Pool
	RedisClient  *redis.Client
}

type Handler struct {
	Ctx *AppContext
}

func NewHandler(ctx *AppContext) *Handler {
	return &Handler{Ctx: ctx}
}

func ReturnHandler(appCtx *AppContext) *ApiHandler {
	return &ApiHandler{
		Logger:       appCtx.Logger,
		RedisClient:  appCtx.RedisClient,
		PostgresPool: appCtx.PostgresPool,
	}
}
