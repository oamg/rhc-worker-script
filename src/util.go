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
	"gopkg.in/yaml.v3"
)

// writeFileToTemporaryDir writes the provided data to a temporary file in the
// designated temporary worker directory. It creates the directory if it doesn't exist.
// The function returns the filename of the created temporary file.
func writeFileToTemporaryDir(data []byte, temporaryWorkerDirectory string) string {
	if err := checkAndCreateDirectory(temporaryWorkerDirectory); err != nil {
		log.Error("Failed to create temporary directory: ", err)
	}

	file, err := os.CreateTemp(temporaryWorkerDirectory, "rhc-worker-script")
	if err != nil {
		log.Errorln("Failed to create temporary file: ", err)
	}

	if _, err := file.Write(data); err != nil {
		log.Errorln("Failed to write content to temporary file: ", err)
	}

	fileName := file.Name()
	if err := file.Close(); err != nil {
		log.Errorln("File was unexpectedly already closed: ", err)
	}
	return fileName
}

// Expected JSON format of message by Insights Upload service (Ingress)
type jsonResponseFormat struct {
	CorrelationID string `json:"correlation_id"`
	Stdout        string `json:"stdout"`
}

// getOutputFile creates a multipart form-data payload for the executed script's output.
// It takes the script file name, stdout string, correlation ID, and content type as input.
// The function constructs the form-data payload containing the script output as a JSON
// file and returns the payload as a *bytes.Buffer and the boundary string.
func getOutputFile(stdout string, correlationID string, contentType string) (*bytes.Buffer, string) {
	payloadData := jsonResponseFormat{CorrelationID: correlationID, Stdout: stdout}
	jsonPayload, err := json.Marshal(payloadData)
	if err != nil {
		log.Errorln("Failed to marshal paylod data: ", err)
	}
	reader := bytes.NewReader(jsonPayload)

	log.Infoln("Writing form-data for executed script: ")
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="%s"; filename="%s"`, "file", "rhc-worker-script-output.tar.gz"))
	h.Set("Content-Type", contentType)

	part, err := writer.CreatePart(h)
	if err != nil {
		log.Errorln("Couldn't create form-file: ", err)
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		log.Errorln("Failed to write json with script stdout to file: ", err)
	}

	if err := writer.Close(); err != nil {
		log.Errorln("Writer was unexpectedly already closed: ", err)
	}

	log.Infoln("form-data created, returning body: ", body)
	return body, writer.Boundary()
}

// constructMetadata constructs a new metadata map by merging the receivedMetadata map
// with an additional "Content-Type" key-value pair. It takes the received metadata map
// and the content type string as input and returns a new metadata map.
func constructMetadata(receivedMetadata map[string]string, contentType string) map[string]string {
	ourMetadata := map[string]string{
		"Content-Type": contentType,
	}
	for k, v := range receivedMetadata {
		ourMetadata[k] = v
	}
	return ourMetadata
}

// Struc used fro worker global config
type Config struct {
	Directive                *string `yaml:"directive,omitempty"`
	VerifyYAML               *bool   `yaml:"verify_yaml,omitempty"`
	InsightsCoreGPGCheck     *bool   `yaml:"insights_core_gpg_check,omitempty"`
	TemporaryWorkerDirectory *string `yaml:"temporary_worker_directory,omitempty"`
}

// Set default values for the Config struct
func setDefaultValues(config *Config) {
	// Set default values for string and boolean fields if they are nil (not present in the YAML)
	if config.Directive == nil {
		defaultDirectiveValue := "rhc-worker-script"
		log.Infof("config 'directive' value is empty default value (%s) will be used", defaultDirectiveValue)
		config.Directive = &defaultDirectiveValue
	}

	if config.VerifyYAML == nil {
		defaultVerifyYamlValue := true
		log.Infof("config 'verify_yaml' value is empty default value (%t) will be used", defaultVerifyYamlValue)
		config.VerifyYAML = &defaultVerifyYamlValue
	}

	if config.InsightsCoreGPGCheck == nil {
		defaultGpgCheckValue := true
		log.Infof("config 'insights_core_gpg_check' value is empty default value (%t) will be used", defaultGpgCheckValue)
		config.InsightsCoreGPGCheck = &defaultGpgCheckValue
	}

	if config.TemporaryWorkerDirectory == nil {
		defaultTemporaryWorkerDirectoryValue := "/var/lib/rhc-worker-script"
		log.Infof("config 'temporary_worker_directory' value is empty default value (%s) will be used", defaultTemporaryWorkerDirectoryValue)
		config.TemporaryWorkerDirectory = &defaultTemporaryWorkerDirectoryValue
	}
}

// Load yaml config, if file doesn't exist or is invalid yaml then empty Config is returned
func loadYAMLConfig(filePath string) *Config {
	var config Config

	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Error(err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		log.Error(err)
	}

	return &config
}

// Load config from given filepath, if config doesn't exist then default config values are used
// Directive = rhc-worker-script
// VerifyYAML = "1"
// InsightsCoreGPGCheck = "1"
func loadConfigOrDefault(filePath string) *Config {
	config := &Config{}
	_, err := os.Stat(filePath)
	if err == nil {
		// File exists, load configuration from YAML
		config = loadYAMLConfig(filePath)
	}

	// File doesn't exist, create a new Config with default values
	setDefaultValues(config)
	return config
}

// Helper function to check if a directory exists, if not, create the
// directory, otherwise, return nil. If it fails to create the directory, the
// function will return the error raised by `os.Mkdir` to the caller.
func checkAndCreateDirectory(folder string) error {
	// Check if path exists, if not, create it.
	if _, err := os.Stat(folder); err != nil {
		// Owner has permission to list, append and search in folder
		if err := os.Mkdir(folder, 0700); err != nil {
			return err
		}
	}

	return nil
}
