package logger

import (
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger
var logFile *os.File

func Init() {
	// Get the user's temporary directory
	tmpDir := os.TempDir()

	// Create the log file path
	logFilePath := filepath.Join(tmpDir, "nats-ui.log")

	// Open the log file
	var err error
	logFile, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("Failed to open log file: %v", err)
	}

	// Initialize the logger
	logger = log.New(logFile, "", log.LstdFlags)
}

func Warn(format string, v ...interface{}) {
	logger.Printf("[WARN] "+format, v...)
}

func Info(format string, v ...interface{}) {
	logger.Printf("[INFO] "+format, v...)
}

func Error(format string, v ...interface{}) {
	logger.Printf("[ERROR] "+format, v...)
}

func Debug(format string, v ...interface{}) {
	logger.Printf("[DEBUG] "+format, v...)
}

func CloseLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
