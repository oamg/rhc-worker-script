package main

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	pb "github.com/redhatinsights/yggdrasil/protocol"
)

func TestCreateDataMessage(t *testing.T) {
	testCases := []struct {
		name                  string
		commandOutput         string
		metadata              map[string]string
		directive             string
		messageID             string
		expectedCorrelationID string
		expectedMessageType   string
		expectedDirective     string
		expectedMessageID     string
	}{
		{
			name:          "commandOutput is not empty",
			commandOutput: "Output of the command",
			metadata: map[string]string{
				"correlation_id":      "123",
				"return_content_type": "application/json",
				"return_url":          "example.com",
			},
			directive:             "Directive value",
			messageID:             "Message ID",
			expectedCorrelationID: "123",
			expectedMessageType:   "multipart/form-data",
			expectedDirective:     "example.com",
			expectedMessageID:     "Message ID",
		},
		{
			name:          "commandOutput is empty",
			commandOutput: "",
			metadata: map[string]string{
				"correlation_id":      "456",
				"return_content_type": "text/plain",
				"return_url":          "example.org",
			},
			directive:             "Another directive",
			messageID:             "Another message ID",
			expectedCorrelationID: "456",
			expectedMessageType:   "multipart/form-data",
			expectedDirective:     "Another directive",
			expectedMessageID:     "Another message ID",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := createDataMessage(tc.commandOutput, tc.metadata, tc.directive, tc.messageID)

			if tc.commandOutput == "" {
				// Checks for commandOutput being empty
				if data.Directive != tc.directive {
					t.Errorf("Expected Directive to be %s, but got %s", tc.directive, data.Directive)
				}
				if string(data.Content) != "" {
					t.Errorf("Expected Content to be empty, but got %s", data.Content)
				}
			} else {
				// Checks for commandOutput being non-empty
				if data.Metadata["correlation_id"] != tc.expectedCorrelationID {
					t.Errorf("Expected correlation_id to be %s, but got %s", tc.expectedCorrelationID, data.Metadata["correlation_id"])
				}

				expectedMessageType := "multipart/form-data"
				if !strings.HasPrefix(data.Metadata["Content-Type"], expectedMessageType) {
					t.Errorf("Expected Content-Type to have prefix %s, but got %s", expectedMessageType, data.Metadata["Content-Type"])
				}

				if data.Directive != tc.expectedDirective {
					t.Errorf("Expected Directive to be %s, but got %s", tc.expectedDirective, data.Directive)
				}

				if data.ResponseTo != tc.expectedMessageID {
					t.Errorf("Expected ResponseTo to be %s, but got %s", tc.expectedMessageID, data.ResponseTo)
				}
			}
		})
	}
}

func TestProcessData(t *testing.T) {
	testCases := []struct {
		name                  string
		yamlData              []byte
		expectedOutput        string
		expectedDirective     string
		expectedReturnContent string
	}{
		{
			name:                  "Expected data are present in result data",
			yamlData:              ExampleYamlData,
			expectedOutput:        "Hello World!",
			expectedDirective:     "bar",
			expectedReturnContent: "foo",
		},
	}

	// Set the mock target address
	yggdDispatchSocketAddr = "mock-target"
	defer func() {
		yggdDispatchSocketAddr = "" // Reset the mock target address after the tests
	}()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shouldVerifyYaml := false
			temporaryWorkerDirectory := t.TempDir()
			config = &Config{
				VerifyYAML:               &shouldVerifyYaml,
				TemporaryWorkerDirectory: &temporaryWorkerDirectory,
			}

			returnURL := "bar"
			testData := &pb.Data{
				Content: tc.yamlData,
				Metadata: map[string]string{
					"return_content_type": "foo",
					"return_url":          returnURL,
					"correlation_id":      "000",
				},
				Directive: "Your directive",
				MessageId: "Your message ID",
			}

			data := processData(testData)

			if !strings.Contains(string(data.GetContent()), tc.expectedOutput) {
				t.Errorf("Expected content to contain '%s', but it didn't", tc.expectedOutput)
			}

			if data.GetDirective() != tc.expectedDirective {
				t.Errorf("Expected directive to contain '%s', but it didn't", tc.expectedDirective)
			}

			if data.GetMetadata()["return_content_type"] != tc.expectedReturnContent {
				t.Errorf("Expected return content type to contain '%s', but it didn't", tc.expectedReturnContent)
			}
		})
	}
}

func TestSendDataToDispatcher(t *testing.T) {
	testCases := []struct {
		name     string
		testData *pb.Data
	}{
		{
			name: "Data are not changed before sending them to dispatcher",
			testData: &pb.Data{
				MessageId:  uuid.New().String(),
				ResponseTo: "mock-id",
			},
		},
	}

	// Set the mock target address
	yggdDispatchSocketAddr = "mock-target"
	defer func() {
		yggdDispatchSocketAddr = "" // Reset the mock target address after the tests
	}()

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			data := sendDataToDispatcher(tc.testData)
			if data != tc.testData {
				t.Errorf("Function should NOT change data before sending, but it did: %s", data)
			}
		})
	}
}
