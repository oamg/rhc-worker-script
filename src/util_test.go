package main

import (
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestConstructMetadata(t *testing.T) {
	testCases := []struct {
		name             string
		receivedMetadata map[string]string
		contentType      string
		expectedMetadata map[string]string
	}{
		{
			name: "Metadata created with expected values",
			receivedMetadata: map[string]string{
				"Key1": "Value1",
				"Key2": "Value2",
			},
			contentType: "application/json",
			expectedMetadata: map[string]string{
				"Content-Type": "application/json",
				"Key1":         "Value1",
				"Key2":         "Value2",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := constructMetadata(tc.receivedMetadata, tc.contentType)

			if !reflect.DeepEqual(result, tc.expectedMetadata) {
				t.Errorf("Unexpected metadata. Expected: %v, but got: %v", tc.expectedMetadata, result)
			}
		})
	}
}

func TestGetOutputFile(t *testing.T) {
	testCases := []struct {
		name          string
		stdout        string
		correlationID string
		contentType   string
	}{
		{
			name:          "Output file contains expected data",
			stdout:        "Hello, world!",
			correlationID: "12345",
			contentType:   "application/json",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			body, boundary := getOutputFile(tc.stdout, tc.correlationID, tc.contentType)

			// Verify that the boundary is not empty
			if boundary == "" {
				t.Error("Boundary should not be empty")
			}

			expectedPayload := fmt.Sprintf(`{"correlation_id":"%s","stdout":"%s"}`, tc.correlationID, tc.stdout)
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
		})
	}
}

func TestWriteFileToTemporaryDir(t *testing.T) {
	testCases := []struct {
		name     string
		data     []byte
		expected string
	}{
		{
			name:     "File is created and contains expected data",
			data:     []byte("test data"),
			expected: "test data",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a temporary directory for testing
			tempDirPath := t.TempDir()

			filePath := writeFileToTemporaryDir(tc.data, tempDirPath)

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
			if string(fileContent) != tc.expected {
				t.Errorf("Expected file content: %s, got: %s", tc.expected, string(fileContent))
			}
		})
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

// Test YAML data
const validYAMLData = `
directive: "rhc-worker-bash"
verify_yaml: true
verify_yaml_version_check: true
insights_core_gpg_check: true
temporary_worker_directory: "/var/lib/rhc-worker-bash"
`

const validYAMLDataMissingValues = `
directive: "rhc-worker-bash"
`

func TestLoadConfigOrDefault(t *testing.T) {
	expectedConfig := &Config{
		Directive:                strPtr("rhc-worker-bash"),
		VerifyYAML:               boolPtr(true),
		InsightsCoreGPGCheck:     boolPtr(true),
		TemporaryWorkerDirectory: strPtr("/var/lib/rhc-worker-bash"),
	}

	testCases := []struct {
		name        string
		yamlData    string
		isValidYAML bool
	}{
		{
			name:        "No config present, defaults set",
			yamlData:    "",
			isValidYAML: false,
		},
		{
			name:        "Valid YAML file with all values present",
			yamlData:    validYAMLData,
			isValidYAML: true,
		},
		{
			name:        "Valid YAML file with missing values",
			yamlData:    validYAMLDataMissingValues,
			isValidYAML: true,
		},
		{
			name:        "Invalid YAML file - default config created",
			yamlData:    "invalid_yaml_data",
			isValidYAML: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			filePath, err := createTempYAMLFile(tc.yamlData)
			if err != nil {
				t.Fatalf("Failed to create temporary YAML file: %v", err)
			}
			defer os.Remove(filePath)

			config := loadConfigOrDefault(filePath)

			// Verify if the YAML is valid and matches the expected config
			if tc.isValidYAML {
				if !compareConfigs(config, expectedConfig) {
					t.Errorf("Loaded config does not match expected config")
				}
			} else {
				// If the YAML is invalid, a default config should be created
				if !compareConfigs(config, expectedConfig) {
					t.Errorf("Loaded config does not match the default config")
				}
			}
		})
	}
}
