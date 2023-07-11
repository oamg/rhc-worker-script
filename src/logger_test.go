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

	// Test case 1: YGG_LOG_LEVEL doesn't exist, info level should be set to info
	setupLogger(logFolderName, logFileName)
	level := log.CurrentLevel()
	if log.CurrentLevel() != log.LevelInfo {
		t.Errorf("Incorrect log level. Expected: %v, Got: %v", log.LevelInfo, level)
	}
	// Test case 2: Unparsable level in env variable
	os.Setenv("YGG_LOG_LEVEL", "....")
	setupLogger(logFolderName, logFileName)
	level = log.CurrentLevel()
	if log.CurrentLevel() != log.LevelInfo {
		t.Errorf("Incorrect log level. Expected: %v, Got: %v", log.LevelInfo, level)
	}

	// Test case 3: Everything set up correctly
	os.Setenv("YGG_LOG_LEVEL", "debug")

	logfile := setupLogger(logFolderName, logFileName)
	if _, err := os.Stat(logFolderName); os.IsNotExist(err) {
		t.Errorf("Log folder not created: %v", err)
	}
	logFilePath := filepath.Join(logFolderName, logFileName)
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("Log file not created: %v", err)
	}
	level = log.CurrentLevel()
	if level != log.LevelDebug {
		t.Errorf("Incorrect log level. Expected: %v, Got: %v", log.LevelDebug, level)
	}
	flags := log.Flags()
	expectedFlags := log.Lshortfile | log.LstdFlags
	if flags != expectedFlags {
		t.Errorf("Incorrect log flags. Expected: %v, Got: %v", expectedFlags, flags)
	}

	// Cleanup - close the log file
	defer os.Unsetenv("YGG_LOG_LEVEL")
	err := logfile.Close()
	if err != nil {
		t.Errorf("Failed to close log file: %v", err)
	}
}
