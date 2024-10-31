package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger
var logFile *os.File

func init() {
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

func fwarn(format string, v ...interface{}) {
	logger.Printf("[WARN] "+format, v...)
}

func finfo(format string, v ...interface{}) {
	logger.Printf("[INFO] "+format, v...)
}

func ferror(format string, v ...interface{}) {
	logger.Printf("[ERROR] "+format, v...)
}

func fdebug(format string, v ...interface{}) {
	logger.Printf("[DEBUG] "+format, v...)
}

func closeLogger() {
	if logFile != nil {
		logFile.Close()
	}
}
