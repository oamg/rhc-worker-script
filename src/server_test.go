package main

import (
	"strings"
	"testing"
)

func TestCreateDataMessage(t *testing.T) {
	// Test case 1: commandOutput is not empty
	commandOutput := "Output of the command"
	metadata := map[string]string{
		"correlation_id":      "123",
		"return_content_type": "application/json",
		"return_url":          "example.com",
	}
	directive := "Directive value"
	messageID := "Message ID"

	data := createDataMessage(commandOutput, metadata, directive, messageID)

	expectedCorrelationID := "123"
	if data.Metadata["correlation_id"] != expectedCorrelationID {
		t.Errorf("Expected correlation_id to be %s, but got %s", expectedCorrelationID, data.Metadata["correlation_id"])
	}

	expectedMessageType := "multipart/form-data"
	if !strings.HasPrefix(data.Metadata["Content-Type"], expectedMessageType) {
		t.Errorf("Expected Content-Type to have prefix %s, but got %s", expectedMessageType, data.Metadata["Content-Type"])
	}

	expectedDirective := "example.com"
	if data.Directive != expectedDirective {
		t.Errorf("Expected Directive to be %s, but got %s", expectedDirective, data.Directive)
	}

	expectedMessageID := messageID
	if data.ResponseTo != expectedMessageID {
		t.Errorf("Expected ResponseTo to be %s, but got %s", expectedMessageID, data.ResponseTo)
	}

	commandOutput = ""
	data = createDataMessage(commandOutput, metadata, directive, messageID)
	if data.Directive != directive {
		t.Errorf("Expected Directive to be %s, but got %s", directive, data.Directive)
	}
	if string(data.Content) != "" {
		t.Errorf("Expected Content to be empty, but got %s", data.Content)
	}
}
