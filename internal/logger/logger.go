package logger

import (
	"log"
	"log/slog"
	"os"
	"path/filepath"
)

func NewLogger() *slog.Logger {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		log.Fatalf("Failed to get home directory: %v", err)
	}
	logDir := filepath.Join(homeDir, ".config/godo/logs")
	os.MkdirAll(logDir, 0755)
	logFile := filepath.Join(logDir, "app.log")
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.FileMode(0644))
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	handler := slog.NewTextHandler(file, nil)
	logger := slog.New(handler)
	return logger
}
