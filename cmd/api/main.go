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
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/logger"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type App struct {
	Config config.Config
	Logger *zap.Logger
	Server *http.Server
}

func main() {
	logger, err := logger.SetupLogger()
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to initialize logger")
		panic(err)
	}
	config := &config.Config{
		ServerPort: "8000",
	}
	defer logger.Sync()

	app := App{Config: *config, Logger: logger}
	app.startServer()
}

func (app *App) startServer() {
	sugaredLogger := logger.FetchSugaredLogger(app.Logger)
	router := mux.NewRouter()
	initializeRoutes(router, app.Logger)

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

	// Listen for termination signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	sugaredLogger.Info("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		sugaredLogger.Panicf("Server Shutdown Failed:%+v", err)
	}
	sugaredLogger.Info("Server stopped successfully.")
}

func initializeRoutes(router *mux.Router, logger *zap.Logger) {
	handler := api.ReturnHandler(logger)

	v1 := router.PathPrefix("/apis/v1").Subrouter()
	v1.HandleFunc("/submit-job", handler.SubmitJob).Methods("POST")

	v1.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

}
