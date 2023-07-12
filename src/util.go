package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"

	"git.sr.ht/~spc/go-log"
)

// Calls os.LookupEnv for key, if not found then fallback value is returned
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func writeFileToTemporaryDir(data []byte, temporaryWorkerDirectory string) string {
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

type jsonResponseFormat struct {
	CorrelationID string `json:"correlation_id"`
	Stdout        string `json:"stdout"`
}

func getOutputFile(scriptFileName string, stdout string, correlationID string, contentType string) (*bytes.Buffer, string) {
	payloadData := jsonResponseFormat{CorrelationID: correlationID, Stdout: stdout}
	jsonPayload, err := json.Marshal(payloadData)
	if err != nil {
		log.Errorln("Failed to marshal paylod data: ", err)
	}
	reader := bytes.NewReader(jsonPayload)

	log.Infoln("Writing form-data for executed script: ", scriptFileName)
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", "rhc-worker-bash-output.tar.gz"))
	h.Set("Content-Type", contentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		log.Errorln("Couldn't create form-file: ", err)
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		log.Errorln("Failed to write json with script stdout to file: ", err)
	}

	writer.Close()

	log.Infoln("form-data created, returning body: ", body)
	return body, writer.Boundary()
}

func constructMetadata(receivedMetadata map[string]string, contentType string) map[string]string {
	ourMetadata := map[string]string{
		"Content-Type": contentType,
	}
	for k, v := range receivedMetadata {
		ourMetadata[k] = v
	}
	return ourMetadata
}
