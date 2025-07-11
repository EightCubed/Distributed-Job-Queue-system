package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/api"
	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/gorilla/mux"
)

type App struct {
	Config config.Config
	Server *http.Server
}

func main() {
	app := App{Config: config.Config{ServerPort: "8000"}}
	app.startServer()
}

func (app *App) startServer() {
	router := mux.NewRouter()
	initializeRoutes(router)

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
		log.Printf("Starting server on port %s...\n", app.Config.ServerPort)
		if err := app.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	// Listen for termination signals
	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
	<-stopChan

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Println("Server stopped successfully.")
}

func initializeRoutes(router *mux.Router) {
	v1 := router.PathPrefix("/apis/v1").Subrouter()
	v1.HandleFunc("/submit-job", api.SubmitJob).Methods("POST")
}
