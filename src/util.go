package main

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"git.sr.ht/~spc/go-log"
)

const temporaryConvert2RHELDir string = "/var/lib/convert2rhel"

func writeFileToTemporaryDir(data []byte) string {
	// Check if path exists, if not, create it.
	if _, err := os.Stat(temporaryConvert2RHELDir); err != nil {
		if err := os.Mkdir(temporaryConvert2RHELDir, os.ModePerm); err != nil {
			log.Errorln(err)
		}
	}

	file, err := os.CreateTemp(temporaryConvert2RHELDir, "c2r-worker-")
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
	output, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Errorln("Couldn't read file")
	}

	if err := json.Valid(output); !err {
		log.Errorln("Can't unmarshal contents of file.")
	}

	return output
}
