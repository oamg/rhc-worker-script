package main

import (
	"strings"
	"testing"

	"github.com/google/uuid"
	pb "github.com/redhatinsights/yggdrasil/protocol"
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

func TestProcessData(t *testing.T) {
	// FIXME: this should ideally test that all correct functions are called
	// Probably easiest would be to move them to interface and then make the function argument take mock interface
	yggdDispatchSocketAddr = "mock-target"

	shouldVerifyYaml := false
	shouldDoInsightsCoreGPGCheck := false
	temporaryWorkerDirectory := "test-dir"
	config = &Config{
		VerifyYAML:               &shouldVerifyYaml,
		TemporaryWorkerDirectory: &temporaryWorkerDirectory,
		InsightsCoreGPGCheck:     &shouldDoInsightsCoreGPGCheck,
	}

	yamlData := []byte(`
vars:
    _insights_signature: "invalid-signature"
    _insights_signature_exclude: "/vars/insights_signature,/vars/content_vars"
    content: |
        #!/bin/sh
        echo "$RHC_WORKER_FOO $RHC_WORKER_BAR!"
    content_vars:
        FOO: Hello
        BAR: World`)

	returnURL := "bar"
	testData := &pb.Data{
		Content: yamlData,
		Metadata: map[string]string{
			"return_content_type": "foo",
			"return_url":          returnURL,
			"correlation_id":      "000",
		},
		Directive: "Your directive",
		MessageId: "Your message ID",
	}

	data := processData(testData)
	expectedOutput := "Hello World!"

	if !strings.Contains(string(data.GetContent()), expectedOutput) {
		t.Errorf("Expected content to contain '%s', but it didn't", expectedOutput)
	}

	if data.GetDirective() != returnURL {
		t.Errorf("Expected directive to contain '%s', but it didn't", returnURL)
	}
}

func TestSendDataToDispatcher(t *testing.T) {
	// Tests only that the function doesn't modify data sent to dispatcher

	yggdDispatchSocketAddr = "mock-target"

	testData := &pb.Data{
		MessageId:  uuid.New().String(),
		ResponseTo: "mock-id",
	}

	data := sendDataToDispatcher(testData)
	if data != testData {
		t.Errorf("Function should NOT change data before sent, but it did: %s", data)
	}
}
