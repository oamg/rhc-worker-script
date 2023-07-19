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
	LogDir                   *string `yaml:"log_dir,omitempty"`
	LogFileName              *string `yaml:"log_filename,omitempty"`
}

// Set default values for the Config struct
func setDefaultValues(config *Config) {
	// Set default values for string and boolean fields if they are nil (not present in the YAML)
	if config.Directive == nil {
		defaultDirectiveValue := "rhc-worker-bash"
		config.Directive = &defaultDirectiveValue
	}

	if config.VerifyYAML == nil {
		defaultVerifyYamlValue := true
		config.VerifyYAML = &defaultVerifyYamlValue
	}

	if config.InsightsCoreGPGCheck == nil {
		defaultGpgCheckValue := true
		config.InsightsCoreGPGCheck = &defaultGpgCheckValue
	}

	if config.TemporaryWorkerDirectory == nil {
		defaultTemporaryWorkerDirectoryValue := "/var/lib/rhc-worker-bash"
		config.TemporaryWorkerDirectory = &defaultTemporaryWorkerDirectoryValue
	}

	if config.LogDir == nil {
		defaultLogFolder := "/var/log/rhc-worker-bash"
		config.LogDir = &defaultLogFolder
	}

	if config.LogFileName == nil {
		defaultLogFilename := "rhc-worker-bash.log"
		config.LogFileName = &defaultLogFilename
	}
}

// Load yaml config, if file doesn't exist or is invalid yaml then empty COnfig is returned
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
// Directive = rhc-worker-bash
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
