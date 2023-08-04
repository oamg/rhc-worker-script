package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"testing"

	"git.sr.ht/~spc/go-log"
)

func TestSetupLogger(t *testing.T) {
	testCases := []struct {
		name          string
		logLevelEnv   string
		expectedLevel log.Level
		expectedFlags int
	}{
		{
			name:          "YGG_LOG_LEVEL doesn't exist, info level should be set to info",
			logLevelEnv:   "",
			expectedLevel: log.LevelInfo,
			expectedFlags: log.Lshortfile | log.LstdFlags,
		},
		{
			name:          "Unparsable level in env variable",
			logLevelEnv:   "....",
			expectedLevel: log.LevelInfo,
			expectedFlags: log.Lshortfile | log.LstdFlags,
		},
		{
			name:          "Everything set up correctly",
			logLevelEnv:   "debug",
			expectedLevel: log.LevelDebug,
			expectedFlags: log.Lshortfile | log.LstdFlags,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory for the log folder
			logFolderName := "log-test"
			logFileName := "log-file"
			defer os.RemoveAll(logFolderName)

			// Set YGG_LOG_LEVEL environment variable
			if tc.logLevelEnv != "" {
				os.Setenv("YGG_LOG_LEVEL", tc.logLevelEnv)
			}

			logfile := setupLogger(logFolderName, logFileName)

			// Verify log level
			level := log.CurrentLevel()
			if level != tc.expectedLevel {
				t.Errorf("Incorrect log level. Expected: %v, Got: %v", tc.expectedLevel, level)
			}

			// Verify log file and folder
			if _, err := os.Stat(logFolderName); os.IsNotExist(err) {
				t.Errorf("Log folder not created: %v", err)
			}
			logFilePath := filepath.Join(logFolderName, logFileName)
			if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
				t.Errorf("Log file not created: %v", err)
			}

			// Verify log flags
			flags := log.Flags()
			if flags != tc.expectedFlags {
				t.Errorf("Incorrect log flags. Expected: %v, Got: %v", tc.expectedFlags, flags)
			}

			// Cleanup - close the log file and unset the YGG_LOG_LEVEL environment variable
			defer func() {
				os.Unsetenv("YGG_LOG_LEVEL")
				err := logfile.Close()
				if err != nil {
					t.Errorf("Failed to close log file: %v", err)
				}
			}()
		})
	}
}

func TestSetupSosExtrasReport(t *testing.T) {
	// FIXME: We are overriding the globals for the below variables, not the
	// best approach, but works for now.
	sosReportFile = "log-file"
	sosReportFolder = t.TempDir()
	fileContent := path.Join(sosReportFolder, sosReportFile, "test-file")
	expectedFileContent := fmt.Sprintf(":%s", fileContent)

	setupSosExtrasReport(fileContent)
	if _, err := os.Stat(sosReportFolder); os.IsNotExist(err) {
		t.Errorf("Log folder not created: %v", err)
	}

	logFilePath := filepath.Join(sosReportFolder, sosReportFile)
	if _, err := os.Stat(logFilePath); os.IsNotExist(err) {
		t.Errorf("SOS report file not created: %v", err)
	}

	logFile, err := os.ReadFile(logFilePath)
	if err != nil {
		t.Errorf("Failed to read file: %v", err)
	}
	if string(logFile) != expectedFileContent {
		t.Errorf("File content does not match")
	}
}
