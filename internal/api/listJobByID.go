package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/EightCubed/Distributed-Job-Queue-system/pkg/models"
	"github.com/gorilla/mux"
)

func (handler *ApiHandler) ListJobByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := config.LoggerFromContext(ctx)
	sugar := logger.Sugar()

	vars := mux.Vars(r)
	jobIDStr := vars["job_id"]
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		sugar.Warnf("Failed to parse request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	sugar.Infof("Listing job-id %d", jobID)

	query := `
		SELECT id, type, data, message, priority, status, created_at, execution_at
		FROM jobs
		WHERE id = $1
	`

	var job models.Job
	err = handler.PostgresPool.QueryRow(ctx, query, jobID).Scan(
		&job.ID,
		&job.Type,
		&job.Data,
		&job.Message,
		&job.Priority,
		&job.Status,
		&job.CreatedAt,
		&job.ExecutionAt,
	)
	if err != nil {
		sugar.Error("Failed to fetch job", err)
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(job); err != nil {
		sugar.Error("Failed to encode job to JSON", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
