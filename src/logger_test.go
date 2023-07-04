package main

import (
	"os"
	"path/filepath"
	"testing"

	"git.sr.ht/~spc/go-log"
)

func TestSetupLogger(t *testing.T) {
	// Create a temporary directory for the log folder
	logFolderName := "log-test"
	logFileName := "log-file"

	defer os.RemoveAll(logFolderName)

	// Mock the YGG_LOG_LEVEL environment variable
	os.Setenv("YGG_LOG_LEVEL", "debug")
	defer os.Unsetenv("YGG_LOG_LEVEL")

	// Call the function being tested
	logfile := setupLogger(logFolderName, logFileName)

	// Verify that the log folder and file were created
	if _, err := os.Stat(logFolderName); os.IsNotExist(err) {
		t.Errorf("Log folder not created: %v", err)
	}

	logFilePath := filepath.Join(logFolderName, logFileName)
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Log file not created: %v", err)
	}

	// Verify that the log level was set correctly
	level := log.CurrentLevel()
	if level != log.LevelDebug {
		t.Errorf("Incorrect log level. Expected: %v, Got: %v", log.LevelDebug, level)
	}

	// Verify that the log flags were set correctly
	flags := log.Flags()
	expectedFlags := log.Lshortfile | log.LstdFlags
	if flags != expectedFlags {
		t.Errorf("Incorrect log flags. Expected: %v, Got: %v", expectedFlags, flags)
	}

	// Cleanup - close the log file
	err := logfile.Close()
	if err != nil {
		t.Errorf("Failed to close log file: %v", err)
	}
}
