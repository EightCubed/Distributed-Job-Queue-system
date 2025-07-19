package api

import (
	"encoding/json"
	"net/http"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/config"
	"github.com/EightCubed/Distributed-Job-Queue-system/pkg/models"
)

func (handler *ApiHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger := config.LoggerFromContext(ctx)
	sugar := logger.Sugar()

	sugar.Info("Listing job submissions")

	stateParam := r.URL.Query().Get("q")
	jobStatus, ok := IsValidJobStatus(stateParam)

	query := "SELECT id, type, data, message, priority, status, created_at, execution_at FROM jobs"
	var args []interface{}

	if ok {
		query += " WHERE status = $1"
		args = append(args, jobStatus)
	}

	rows, err := handler.PostgresPool.Query(ctx, query, args...)
	if err != nil {
		sugar.Error("Failed to fetch jobs", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var jobs []models.Job
	for rows.Next() {
		var job models.Job
		if err := rows.Scan(
			&job.ID, &job.Type, &job.Data, &job.Message, &job.Priority,
			&job.Status, &job.CreatedAt, &job.ExecutionAt,
		); err != nil {
			sugar.Error("Failed to scan job row", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		jobs = append(jobs, job)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func IsValidJobStatus(s string) (models.JOB_STATUS, bool) {
	status := models.JOB_STATUS(s)
	switch status {
	case models.JOB_STATUS_QUEUED, models.JOB_STATUS_PROGRESS,
		models.JOB_STATUS_COMPLETED, models.JOB_STATUS_FAILED:
		return status, true
	default:
		return "", false
	}
}
