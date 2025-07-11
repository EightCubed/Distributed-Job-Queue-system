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

// Nicer looking logs but slower
func FetchSugaredLogger(logger *zap.Logger) *zap.SugaredLogger {
	return logger.Sugar()
}

// Logging for time-bound tasks
func FetchLogger(logger *zap.Logger) *zap.Logger {
	return logger
}
