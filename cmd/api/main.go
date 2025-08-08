package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/api"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/db"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/logger"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	Config       config.Config
	Logger       *zap.Logger
	Server       *http.Server
	RedisClient  *redis.Client
	PostgresPool *pgxpool.Pool
}

type AppContext struct {
	Logger       *zap.Logger
	PostgresPool *pgxpool.Pool
	RedisClient  *redis.Client
}

func main() {
	logger, err := logger.SetupLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger")
		panic(err)
	}

	config := &config.Config{
		ServerPort:    "8000",
		RedisAddr:     "localhost:6379",
		RedisPassword: "",
		PostgresDSN:   "postgres://rockon:password@localhost:5432/job-queue?sslmode=disable",
	}
	defer logger.Sync()

	redisClient := db.NewRedisClient(config.RedisAddr, config.RedisPassword)
	postgresPool, err := db.NewPostgresPool(config.PostgresDSN)
	if err != nil {
		logger.Fatal("Failed to connect to Postgres", zap.Error(err))
	}
	defer postgresPool.Close()

	app := App{Config: *config, Logger: logger, RedisClient: redisClient, PostgresPool: postgresPool}
	app.startServer()
}

func (app *App) startServer() {
	sugaredLogger := logger.FetchSugaredLogger(app.Logger)
	router := mux.NewRouter()
	router.Use(LoggingMiddleware(app.Logger))

	initializeRoutes(router, app)

	app.Server = &http.Server{
		Addr:              ":" + app.Config.ServerPort,
		Handler:           router,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
		MaxHeaderBytes:    1 << 20,
		ReadHeaderTimeout: 2 * time.Second,
	}

	go func() {
		sugaredLogger.Infof("Starting server on port %s...", app.Config.ServerPort)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			sugaredLogger.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	sugaredLogger.Info("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer app.RedisClient.Close()

	if err := app.Server.Shutdown(ctx); err != nil {
		sugaredLogger.Panicf("Server Shutdown Failed:%+v", err)
	}

	sugaredLogger.Info("Server stopped successfully.")
}

func initializeRoutes(router *mux.Router, app *App) {
	appCtx := &api.AppContext{
		Logger:       app.Logger,
		PostgresPool: app.PostgresPool,
		RedisClient:  app.RedisClient,
	}
	handler := api.ReturnHandler(appCtx)

	v1 := router.PathPrefix("/apis/v1").Subrouter()
	v1.HandleFunc("/submit-job", handler.SubmitJob).Methods("POST")
	v1.HandleFunc("/jobs", handler.ListJobs).Methods("GET")
	v1.HandleFunc("/job/{job_id}", handler.ListJobByID).Methods("GET")

	v1.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func LoggingMiddleware(baseLogger *zap.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := uuid.New().String()
			logger := baseLogger.With(zap.String("request_id", reqID))

			ctx := context.WithValue(r.Context(), config.LoggerKey, logger)

			logger.Info("Incoming request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.String("remote", r.RemoteAddr),
			)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
