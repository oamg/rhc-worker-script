package main

import (
	"io"
	"os"
	"path"

	"git.sr.ht/~spc/go-log"
)

const logFolder string = "/var/log/convert2rhel"
const fileName string = "convert2rhel-worker.log"

func setupLogger() *os.File {
	// Check if path exists, if not, create it.
	if _, err := os.Stat(logFolder); err != nil {
		if err := os.Mkdir(logFolder, os.ModePerm); err != nil {
			log.Fatal(err)
		}
	}

	// open log file
	logFile, err := os.Create(path.Join(logFolder, fileName))
	if err != nil {
		log.Fatal(err)
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
