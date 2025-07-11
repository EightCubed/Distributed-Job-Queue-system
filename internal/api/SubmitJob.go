package api

import (
	"encoding/json"
	"net/http"

	"github.com/EightCubed/Distributed-Job-Queue-system/internal/logger"
)

type SubmitJobBody struct {
	Type     string `json:"type"`
	Priority string `json:"priority"`
}

func (handler *ApiHandler) SubmitJob(w http.ResponseWriter, r *http.Request) {
	sugaredLogger := logger.FetchSugaredLogger(handler.Logger)
	sugaredLogger.Info("SubmitJob handler called")

	var body SubmitJobBody
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil {
		sugaredLogger.Info(err)
		return
	}

	sugaredLogger.Info(body)
}
