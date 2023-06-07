package main

import (
	"encoding/json"
	"os"

	"git.sr.ht/~spc/go-log"
)

func writeFileToTemporaryDir(data []byte) string {
	const temporaryWorkerDirectory string = "/var/lib/rhc-bash-worker"

	// Check if path exists, if not, create it.
	if _, err := os.Stat(temporaryWorkerDirectory); err != nil {
		if err := os.Mkdir(temporaryWorkerDirectory, os.ModePerm); err != nil {
			log.Errorln(err)
		}
	}

	file, err := os.CreateTemp(temporaryWorkerDirectory, "c2r-worker-")
	if err != nil {
		log.Errorln(err)
	}

	if _, err := file.Write(data); err != nil {
		log.Errorln(err)
	}

	fileName := file.Name()
	file.Close()
	return fileName
}

func readOutputFile(filePath string) []byte {
	output, err := os.ReadFile(filePath)
	if err != nil {
		log.Errorln("Couldn't read file")
	}

	if err := json.Valid(output); !err {
		log.Errorln("Can't unmarshal contents of file.")
	}

	return output
}
