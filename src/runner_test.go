package main

import (
	"os"
	"os/exec"
	"testing"
)

func TestProcessSignedScript(t *testing.T) {
	testCases := []struct {
		name           string
		verifyYAML     bool
		yamlData       []byte
		expectedResult string
	}{
		{
			name:           "verification disabled, no yaml data supplied = empty output",
			verifyYAML:     false,
			yamlData:       []byte{},
			expectedResult: "",
		},
		{
			name:       "verification disabled, yaml data supplied = non-empty output",
			verifyYAML: false,
			yamlData: []byte(`
vars:
    insights_signature: "invalid-signature"
    insights_signature_exclude: "/vars/insights_signature,/vars/content_vars"
    content: |
        #!/bin/sh
        echo "$RHC_WORKER_FOO $RHC_WORKER_BAR!"
    content_vars:
        FOO: Hello
        BAR: World`),
			expectedResult: "Hello World!\n",
		},
		{
			name:       "verification enabled, invalid signature = error msg returned",
			verifyYAML: true,
			yamlData: []byte(`
vars:
    insights_signature: "invalid-signature"
    insights_signature_exclude: "/vars/insights_signature,/vars/content_vars"
    content: |
        #!/bin/sh
        echo "$RHC_WORKER_FOO $RHC_WORKER_BAR!"
    content_vars:
        FOO: Hello
        BAR: World`),
			expectedResult: "Signature of yaml file is invalid",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shouldVerifyYaml := tc.verifyYAML
			shouldDoInsightsCoreGPGCheck := tc.verifyYAML // Assume the same value for simplicity
			temporaryWorkerDirectory := t.TempDir()
			config = &Config{
				VerifyYAML:               &shouldVerifyYaml,
				TemporaryWorkerDirectory: &temporaryWorkerDirectory,
				InsightsCoreGPGCheck:     &shouldDoInsightsCoreGPGCheck,
			}

			defer os.RemoveAll(temporaryWorkerDirectory)

			result := processSignedScript(tc.yamlData)
			if result != tc.expectedResult {
				t.Errorf("Expected %q, but got %q", tc.expectedResult, result)
			}
		})
	}
}

func TestVerifyYamlFile(t *testing.T) {
	testCases := []struct {
		name                         string
		yamlData                     []byte
		verifyYAML                   bool
		verificationCommand          string
		verificationArgs             []string
		shouldDoInsightsCoreGPGCheck bool
		expectedResult               bool
	}{
		{
			name:                         "verification disabled",
			verifyYAML:                   false,
			yamlData:                     []byte{},
			shouldDoInsightsCoreGPGCheck: false,
			expectedResult:               true,
		},
		{
			name:                         "verification enabled and verification succeeds",
			verifyYAML:                   true,
			yamlData:                     []byte("valid-yaml"),
			verificationCommand:          "true",
			verificationArgs:             []string{},
			shouldDoInsightsCoreGPGCheck: false,
			expectedResult:               true,
		},
		{
			name:                         "verification is enabled and verification fails",
			verifyYAML:                   true,
			yamlData:                     []byte("invalid-yaml"),
			verificationCommand:          "false",
			verificationArgs:             []string{},
			shouldDoInsightsCoreGPGCheck: false,
			expectedResult:               false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			shouldVerifyYaml := tc.verifyYAML
			shouldDoInsightsCoreGPGCheck := tc.shouldDoInsightsCoreGPGCheck
			verificationCommand = tc.verificationCommand
			verificationArgs = tc.verificationArgs

			config = &Config{
				VerifyYAML:           &shouldVerifyYaml,
				InsightsCoreGPGCheck: &shouldDoInsightsCoreGPGCheck,
			}

			result := verifyYamlFile(tc.yamlData)
			if result != tc.expectedResult {
				t.Errorf("Expected %v, but got %v", tc.expectedResult, result)
			}
		})
	}
}

// Function to check if one string slice is a subset of another
// Simple compare isn't enough because environment can change during execution
func areStringSlicesSubset(subset, full []string) bool {
	for _, s := range subset {
		found := false
		for _, f := range full {
			if s == f {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}

func TestSetEnvVariablesForCommand(t *testing.T) {
	testCases := []struct {
		name                string
		variables           map[string]string
		expected            []string
		anotherCmdVariables map[string]string // Same variables with different values for another command
	}{
		{
			name: "SettingVariables",
			variables: map[string]string{
				"VAR1": "value1",
				"VAR2": "value2",
			},
			expected: []string{
				"RHC_WORKER_VAR1=value1",
				"RHC_WORKER_VAR2=value2",
			},
			anotherCmdVariables: map[string]string{
				"VAR1": "another_value1",
			},
		},
		{
			name:      "EmptyVariables",
			variables: nil,
			expected:  nil, // Expect no changes to command's environment in this case
			anotherCmdVariables: map[string]string{
				"VAR2": "another_value2",
			},
		},
		// Add more test cases as needed
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			originalEnv := os.Environ()
			// Create the first dummy command
			cmd := exec.Command("echo", "Hello, World!")
			anotherCmd := exec.Command("echo", "Bye, World!")

			// Call the functions to set env variables for the commands
			setEnvVariablesForCommand(cmd, tc.variables)
			setEnvVariablesForCommand(anotherCmd, tc.anotherCmdVariables)

			// Check if the global environment variables are unchanged
			if !areStringSlicesSubset(originalEnv, os.Environ()) {
				t.Error("Global environment variables have been modified.")
			}

			// Check if the first command's environment variables have been set correctly
			if !areStringSlicesSubset(cmd.Env, append(os.Environ(), tc.expected...)) {
				t.Errorf("Command's environment variables are incorrect. Got: %v, Expected: %v", cmd.Env, append(os.Environ(), tc.expected...))
			}

			// Check if the second command's environment variables are NOT same as for first command
			if areStringSlicesSubset(anotherCmd.Env, append(os.Environ(), tc.expected...)) {
				t.Errorf("Command's environment variables are incorrect. Got: %v, Expected: %v", cmd.Env, append(os.Environ(), tc.expected...))
			}
		})
	}
}
