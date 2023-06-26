package main

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
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

func readOutputFile(filePath string) (*bytes.Buffer, string) {
	log.Infoln("Reading file at:", filePath)
	file, err := os.Open(filePath)
	if err != nil {
		log.Infoln("Failed to read output file: ", err)
		return nil, ""
	}

	log.Infoln("Writing form-data for file: ", filePath)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", "convert2rhel-report.json.tar.gz"))
	h.Set("Content-Type", "application/vnd.redhat.tasks.filename+tgz")
	part, err := writer.CreatePart(h)
	if err != nil {
		log.Errorln("Couldn't create form-file: ", err)
	}
	_, err = io.Copy(part, file)
	if err != nil {
		log.Errorln("Failed to copy contents to file: ", err)
	}

	writer.Close()

	log.Infoln("form-data created, returning body: ", body)
	return body, writer.Boundary()
}
