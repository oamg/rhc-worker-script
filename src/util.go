package main

import (
	"encoding/json"
	"os"

	"git.sr.ht/~spc/go-log"
)

const temporaryWorkerDirectory string = "/var/lib/rhc-worker-bash"

func writeFileToTemporaryDir(data []byte) string {
	// Check if path exists, if not, create it.
	if _, err := os.Stat(temporaryWorkerDirectory); err != nil {
		if err := os.Mkdir(temporaryWorkerDirectory, os.ModePerm); err != nil {
			log.Errorln("Failed to create temporary directory: ", err)
		}
	}

	file, err := os.CreateTemp(temporaryWorkerDirectory, "rhc-worker-bash-")
	if err != nil {
		log.Errorln("Failed to create temporary file: ", err)
	}

	if _, err := file.Write(data); err != nil {
		log.Errorln("Failed to write content to temporary file: ", err)
	}

	fileName := file.Name()
	file.Close()
	return fileName
}

func readOutputFile(filePath string) []byte {
	output, err := os.ReadFile(filePath)
	if err != nil {
		log.Errorln("Failed to read output file: ", err)
		return nil
	}

	if err := json.Valid(output); !err {
		log.Errorln("JSON content is not valid.")
		return nil
	}

	return output
}
