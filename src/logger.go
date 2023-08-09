package main

import (
	"fmt"
	"io"
	"os"
	"path"

	"git.sr.ht/~spc/go-log"
)

var sosReportFolder = "/etc/sos.extras.d"
var sosReportFile = "rhc-worker-script-logs"

// SetupLogger sets up the logger for the application and returns the log file.
// It creates a log folder if it doesn't exist, opens a log file, sets the log level
// based on the "YGG_LOG_LEVEL" environment variable, configures the log output to
// write to both standard output and the log file, and enables optional log features
// such as date-time, filename, and line number.
// Returns a pointer to an os.File representing the opened log file.
func setupLogger(logFolder string, fileName string) *os.File {
	if err := checkAndCreateDirectory(logFolder); err != nil {
		log.Error(err)
	}

	logFilePath := path.Join(logFolder, fileName)
	// open log file
	logFile, err := os.Create(logFilePath)
	if err != nil {
		log.Error(err)
	}

	yggdLogLevel, ok := os.LookupEnv("YGG_LOG_LEVEL")
	if ok {
		level, ok := log.ParseLevel(yggdLogLevel)
		if ok != nil {
			log.Errorf("Could not parse log level '%v', setting the level to info", yggdLogLevel)
			log.SetLevel(log.LevelInfo)
		} else {
			log.SetLevel(level)
		}
	} else {
		// Yggdrasil < 3.0 does not share its configured log level with the
		// workers in any way
		log.SetLevel(log.LevelInfo)
	}

	// Initialization for the sosreport extras plugin
	setupSosExtrasReport(logFilePath)

	// set log output
	multWriter := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(multWriter)

	// optional: log date-time, filename, and line number
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	return logFile
}

// setupSosExtrasReport sets up the sos report file for the sos_extra plugin to
// collect the logs for the worker, which is a special file that points out to
// the current path of the logfile for the worker.
func setupSosExtrasReport(logFilePath string) {
	if err := checkAndCreateDirectory(sosReportFolder); err != nil {
		log.Error(err)
	}

	logFile, err := os.Create(path.Join(sosReportFolder, sosReportFile))
	if err != nil {
		log.Error(err)
	}
	defer logFile.Close()

	// sosreport expects that the file content will be in the following format:
	// ":/path/to/your/log/file.{log,txt,...}", this will trigger sosreport to
	// collect the file without the need to have a special plugin in sosreport.
	content := fmt.Sprintf(":%s", logFilePath)
	if _, err := logFile.WriteString(content); err != nil {
		log.Error(err)
	}
}
