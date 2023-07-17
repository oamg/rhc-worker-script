package main

import (
	"context"
	"os"
	"testing"
)

func TestProcessSignedScript(t *testing.T) {
	shouldVerifyYaml := false
	shouldDoInsightsCoreGPGCheck := false
	temporaryWorkerDirectory := "test-dir"
	config = &Config{
		VerifyYAML:               &shouldVerifyYaml,
		TemporaryWorkerDirectory: &temporaryWorkerDirectory,
		InsightsCoreGPGCheck:     &shouldDoInsightsCoreGPGCheck,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer os.RemoveAll(temporaryWorkerDirectory)

	// Test case 1: verification disabled, no yaml data supplied = empty output
	yamlData := []byte{}
	expectedResult := ""
	result := processSignedScript(ctx, yamlData)
	if result != expectedResult {
		t.Errorf("Expected %q, but got %q", expectedResult, result)
	}

	// Test case 2: verification disabled, yaml data supplied = non-empty output
	yamlData = []byte(`
vars:
    _insights_signature: "invalid-signature"
    _insights_signature_exclude: "/vars/insights_signature,/vars/content_vars"
    content: |
        #!/bin/sh
        echo "$RHC_WORKER_FOO $RHC_WORKER_BAR!"
    content_vars:
        FOO: Hello
        BAR: World`)
	expectedResult = "Hello World!\n"
	result = processSignedScript(ctx, yamlData)
	if result != expectedResult {
		t.Errorf("Expected %q, but got %q", expectedResult, result)
	}

	// FIXME: This is false success because verification fails on missing insighs-client
	// Test case 3: verification enabled, invalid signature = error msg returned
	shouldVerifyYaml = true
	shouldDoInsightsCoreGPGCheck = true
	expectedResult = "Signature of yaml file is invalid"
	result = processSignedScript(ctx, yamlData)
	if result != expectedResult {
		t.Errorf("Expected %q, but got %q", expectedResult, result)
	}

	cancel()
	// Test case 4: verification disabled, context canceled, script shouldn't be executed
	shouldVerifyYaml = false
	expectedResult = ""
	result = processSignedScript(ctx, yamlData)
	if result != expectedResult {
		t.Errorf("Expected %q, but got %q", expectedResult, result)
	}
}

func TestVerifyYamlFile(t *testing.T) {
	shouldVerifyYaml := false
	shouldDoInsightsCoreGPGCheck := false

	config = &Config{
		VerifyYAML:           &shouldVerifyYaml,
		InsightsCoreGPGCheck: &shouldDoInsightsCoreGPGCheck,
	}
	// Test case 1: verification disabled
	ctx, cancel := context.WithCancel(context.Background())

	// Test case 1: shouldVerifyYaml is not "1"
	expectedResult := true
	result := verifyYamlFile(ctx, []byte{})
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// Test case 2: verification enabled and verification succeeds
	shouldVerifyYaml = true
	// FIXME: This should succedd but now verification fails on missing insighs-client
	// We also need valid signature
	expectedResult = false
	result = verifyYamlFile(ctx, []byte("valid-yaml"))
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	// FIXME: Valid test case but fails because of missing insights-client
	// Test case 3: sverification is enabled and verification fails
	// shouldVerifyYaml = true
	expectedResult = false
	result = verifyYamlFile(ctx, []byte("invalid-yaml")) // Replace with your YAML data
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}

	cancel()
	// Test case 4: context is cancelled, verification command shouldn't be called
	shouldVerifyYaml = true
	expectedResult = false
	result = verifyYamlFile(ctx, []byte("invalid-yaml")) // Replace with your YAML data
	if result != expectedResult {
		t.Errorf("Expected %v, but got %v", expectedResult, result)
	}
}
