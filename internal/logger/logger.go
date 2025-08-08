package logger

import (
	"go.uber.org/zap"
)

func SetupLogger() (*zap.Logger, error) {
	logger, err := zap.NewProduction()
	if err != nil {
		return nil, err
	}
	return logger, nil
}

func FetchSugaredLogger(logger *zap.Logger) *zap.SugaredLogger {
	return logger.Sugar()
}

func FetchLogger(logger *zap.Logger) *zap.Logger {
	return logger
}
