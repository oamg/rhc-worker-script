package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestConstructMetadata(t *testing.T) {
	receivedMetadata := map[string]string{
		"Key1": "Value1",
		"Key2": "Value2",
	}
	contentType := "application/json"

	expectedMetadata := map[string]string{
		"Content-Type": "application/json",
		"Key1":         "Value1",
		"Key2":         "Value2",
	}

	result := constructMetadata(receivedMetadata, contentType)

	if !reflect.DeepEqual(result, expectedMetadata) {
		t.Errorf("Unexpected metadata. Expected: %v, but got: %v", expectedMetadata, result)
	}
}

func TestGetOutputFile(t *testing.T) {
	stdout := "Hello, world!"
	correlationID := "12345"
	contentType := "application/json"

	body, boundary := getOutputFile(stdout, correlationID, contentType)

	// Verify that the boundary is not empty
	if boundary == "" {
		t.Error("Boundary should not be empty")
	}

	expectedPayload := fmt.Sprintf(`{"correlation_id":"%s","stdout":"%s"}`, correlationID, stdout)
	gotTrimmed := strings.TrimSpace(body.String())

	// Verify that the body contains the expected data
	if !strings.Contains(gotTrimmed, expectedPayload) {
		t.Errorf("Unexpected body payload. Expected to contain: %s, Got: %s", expectedPayload, body.String())
	}
	prefix := fmt.Sprintf(`--%s`, boundary)
	if !strings.HasPrefix(gotTrimmed, prefix) {
		t.Errorf("Unexpected body payload. Expected to have prefix: %s, Got: %s", prefix, body.String())
	}
	suffix := fmt.Sprintf(`--%s--`, boundary)
	if !strings.HasSuffix(gotTrimmed, suffix) {
		t.Errorf("Unexpected body payload. Expected to have suffix: %s, Got: %s", suffix, gotTrimmed)
	}
	// TODO: test that content type is also there
}

func TestWriteFileToTemporaryDir(t *testing.T) {
	// Create a temporary directory for testing
	tempDirPath := "test-dir"
	defer os.RemoveAll(tempDirPath)

	data := []byte("test data")
	filePath := writeFileToTemporaryDir(data, tempDirPath)

	// Assert that the file exists
	_, err := os.Stat(filePath)
	if err != nil {
		t.Errorf("Expected file to be created, got error: %v", err)
	}

	// Assert that the file contains the expected data
	fileContent, err := os.ReadFile(filePath)
	if err != nil {
		t.Errorf("Failed to read file content: %v", err)
	}
	if string(fileContent) != string(data) {
		t.Errorf("Expected file content: %s, got: %s", string(data), string(fileContent))
	}
}

// Helper function to create a temporary YAML file with the given content and return its path.
func createTempYAMLFile(content string) (string, error) {
	tempFile, err := os.CreateTemp("", "config_test_*.yaml")
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	if _, err := tempFile.WriteString(content); err != nil {
		return "", err
	}

	return tempFile.Name(), nil
}

func TestLoadConfigOrDefault(t *testing.T) {
	expectedConfig := &Config{
		Directive:                strPtr("rhc-worker-bash"),
		VerifyYAML:               boolPtr(true),
		InsightsCoreGPGCheck:     boolPtr(true),
		TemporaryWorkerDirectory: strPtr("/var/lib/rhc-worker-bash"),
	}
	// Test case 1: No config present, defaults set
	config := loadConfigOrDefault("foo-bar")

	if !compareConfigs(config, expectedConfig) {
		t.Errorf("Loaded config does not match expected config")
	}

	// Test case 2: Valid YAML file with all values present
	yamlData := `
directive: "rhc-worker-bash"
verify_yaml: true
verify_yaml_version_check: true
insights_core_gpg_check: true
temporary_worker_directory: "/var/lib/rhc-worker-bash"
`
	filePath, err := createTempYAMLFile(yamlData)
	if err != nil {
		t.Fatalf("Failed to create temporary YAML file: %v", err)
	}
	defer os.Remove(filePath)

	config = loadConfigOrDefault(filePath)

	if !compareConfigs(config, expectedConfig) {
		t.Errorf("Loaded config does not match expected config")
	}

	// Test case 3: Valid YAML file with missing values
	yamlData = `
directive: "rhc-worker-bash"
`
	filePath, err = createTempYAMLFile(yamlData)
	if err != nil {
		t.Fatalf("Failed to create temporary YAML file: %v", err)
	}
	defer os.Remove(filePath)

	config = loadConfigOrDefault(filePath)

	if !compareConfigs(config, expectedConfig) {
		t.Errorf("Loaded config does not match expected config")
	}

	// Test case 4: Invalid YAML file - default config created
	yamlData = `
invalid_yaml_data
`
	filePath, err = createTempYAMLFile(yamlData)
	if err != nil {
		t.Fatalf("Failed to create temporary YAML file: %v", err)
	}
	defer os.Remove(filePath)

	config = loadConfigOrDefault(filePath)

	if !compareConfigs(config, expectedConfig) {
		t.Errorf("Loaded config does not match expected config")
	}
}

// Helper function to compare two Config structs.
func compareConfigs(c1, c2 *Config) bool {
	return *c1.Directive == *c2.Directive &&
		*c1.VerifyYAML == *c2.VerifyYAML &&
		*c1.InsightsCoreGPGCheck == *c2.InsightsCoreGPGCheck &&
		*c1.TemporaryWorkerDirectory == *c2.TemporaryWorkerDirectory
}

// Helper functions for creating pointers to string and bool values.
func strPtr(s string) *string {
	return &s
}

func boolPtr(b bool) *bool {
	return &b
}
