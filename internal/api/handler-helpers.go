package api

import "go.uber.org/zap"

type ApiHandler struct {
	Logger *zap.Logger
}

func ReturnHandler(logger *zap.Logger) *ApiHandler {
	return &ApiHandler{
		Logger: logger,
	}
}
