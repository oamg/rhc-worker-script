package main

import (
	"io"
	"os"
	"path"

	"git.sr.ht/~spc/go-log"
)

// SetupLogger sets up the logger for the application and returns the log file.
// It creates a log folder if it doesn't exist, opens a log file, sets the log level
// based on the "YGG_LOG_LEVEL" environment variable, configures the log output to
// write to both standard output and the log file, and enables optional log features
// such as date-time, filename, and line number.
// Returns a pointer to an os.File representing the opened log file.
func setupLogger(logFolder string, fileName string) *os.File {
	// Check if path exists, if not, create it.
	if _, err := os.Stat(logFolder); err != nil {
		if err := os.Mkdir(logFolder, os.ModePerm); err != nil {
			log.Error(err)
		}
	}

	// open log file
	logFile, err := os.Create(path.Join(logFolder, fileName))
	if err != nil {
		log.Error(err)
	}

	yggdLogLevel, ok := os.LookupEnv("YGG_LOG_LEVEL")
	if ok {
		level, ok := log.ParseLevel(yggdLogLevel)
		if ok != nil {
			log.Errorf("Could not parse log level '%v'", yggdLogLevel)
		} else {
			log.SetLevel(level)
		}
	} else {
		// Yggdrasil < 3.0 does not share its configured log level with the
		// workers in any way
		log.SetLevel(log.LevelInfo)
	}

	// set log output
	multWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multWriter)

	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	return logFile
}
