package main

import (
	"context"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	pb "github.com/redhatinsights/yggdrasil/protocol"
	"github.com/stretchr/testify/assert"
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

func TestListenForTerminationSignal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	sigCh := make(chan os.Signal, 1)

	go listenForTerminationSignal(sigCh, cancel)

	// Simulate the OS signal by sending SIGINT (Ctrl+C)
	sigCh <- syscall.SIGINT

	// Wait for a short duration to allow the goroutine to handle the signal and cancel the context
	time.Sleep(100 * time.Millisecond)

	// Check if the context has been canceled as expected
	assert.True(t, ctx.Err() != nil, "Context should be canceled")
}

func TestProcessData(t *testing.T) {
	// FIXME: this should ideally test that all correct functions are called
	// Probably easiest would be to move them to interface and then make the function argument take mock interface
	yggdDispatchSocketAddr = "mock-target"
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	testData := &pb.Data{
		Content: []byte("Your YAML content"),
		Metadata: map[string]string{
			"return_content_type": "foo",
			"return_url":          "bar",
			"correlation_id":      "000",
		},
		Directive: "Your directive",
		MessageId: "Your message ID",
	}

	processData(ctx, cancel, testData)

	// Check if the context has been canceled as expected
	assert.True(t, ctx.Err() != nil, "Context should be canceled")
}
