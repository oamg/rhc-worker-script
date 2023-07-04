package main

import (
	"os"
	"testing"
)

// NOTE: use fstest instead or regular write?
func TestExecuteScriptSuccess(t *testing.T) {
	// Create a temporary test script file for successful execution
	successScriptFile := "success_script.sh"
	successScriptContent := "#!/bin/sh\necho \"Hello, World!\""
	err := os.WriteFile(successScriptFile, []byte(successScriptContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test script file: %s", err)
	}
	defer os.Remove(successScriptFile)

	// Test successful execution
	successOutput := executeScript(successScriptFile)
	expectedSuccessOutput := "Hello, World!\n"
	if successOutput != expectedSuccessOutput {
		t.Errorf("Unexpected success output. Expected: %q, Got: %q", expectedSuccessOutput, successOutput)
	}
}

// NOTE: use fstest instead of regular write?
func TestExecuteScriptFail(t *testing.T) {
	scriptFile := "foo_bar.sh"

	// Test failed execution - file doesn't exist
	failedOutput := executeScript(scriptFile)
	expectedFailedOutput := ""
	if failedOutput != expectedFailedOutput {
		t.Errorf("Unexpected success output. Expected: %q, Got: %q", expectedFailedOutput, failedOutput)
	}
}

func TestExecuteScriptNoStdout(t *testing.T) {
	// Create a temporary test script file for successful execution
	emptyStdoutScriptFile := "empty_output_script.sh"
	emptyStdoutScriptContent := "#!/bin/sh\necho \"Hello, World!\"> /dev/null"
	err := os.WriteFile(emptyStdoutScriptFile, []byte(emptyStdoutScriptContent), 0600)
	if err != nil {
		t.Fatalf("Failed to create test script file: %s", err)
	}
	defer os.Remove(emptyStdoutScriptFile)

	// Test successful execution
	successOutput := executeScript(emptyStdoutScriptFile)
	expectedSuccessOutput := ""
	if successOutput != expectedSuccessOutput {
		t.Errorf("Unexpected success output. Expected: %q, Got: %q", expectedSuccessOutput, successOutput)
	}
}
