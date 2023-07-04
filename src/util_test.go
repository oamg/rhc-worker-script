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
	scriptFileName := "script.sh"
	stdout := "Hello, world!"
	correlationID := "12345"
	contentType := "application/json"

	body, boundary := getOutputFile(scriptFileName, stdout, correlationID, contentType)

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

func TestGetEnv(t *testing.T) {
	// Test case 1: When the environment variable exists
	key := "MY_VARIABLE"
	fallback := "default"
	expected := "my-value"
	os.Setenv(key, expected)
	defer os.Unsetenv(key)

	result := getEnv(key, fallback)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}

	// Test case 2: When the environment variable does not exist
	key = "NON_EXISTENT_VARIABLE"
	expected = fallback

	result = getEnv(key, fallback)
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}
