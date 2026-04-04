package logger

import (
	"go.uber.org/zap"
)

// NewLogger ...
func NewLogger() *zap.Logger {
	logger, _ := zap.NewProduction()
	return logger
}
