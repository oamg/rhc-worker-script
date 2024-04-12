package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git.sr.ht/~spc/go-log"
	"gopkg.in/yaml.v3"
)

// Nested variables expected to be contained in signedYamlContent
type signedYamlContentVars struct {
	InsightsSignature        string            `yaml:"insights_signature"`
	InsightsSignatureExclude string            `yaml:"insights_signature_exclude"`
	Content                  string            `yaml:"content"`
	Interpreter              string            `yaml:"interpreter"`
	ContentVars              map[string]string `yaml:"content_vars"`
}

// Represents one item in received yaml file
// It is expected that the received yaml file contains list on top level
type signedYamlContent struct {
	Name string                `yaml:"name"`
	Vars signedYamlContentVars `yaml:"vars"`
}

var verificationCommand = "insights-client"

var verificationArgs = []string{
	"-m", "insights.client.apps.ansible.playbook_verifier",
	"--quiet", "--payload", "noop", "--content-type", "noop",
}

// Verify that no one tampered with yaml file
func verifyYamlFile(yamlData []byte) bool {
	if !*config.VerifyYAML {
		log.Warnln("WARNING: Playbook verification disabled.")
		return true
	}

	env := os.Environ()
	log.Infoln("Calling insights-client playbook verifier ...")

	cmd := exec.Command(verificationCommand, verificationArgs...)
	cmd.Env = env
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Errorln(err)
		return false
	}

	// Send yaml data to the command's stdin
	_, err = stdin.Write(yamlData)
	if err != nil {
		log.Errorln(err)
		return false
	}

	if err := stdin.Close(); err != nil {
		log.Errorln("stdin was unexpectedly already closed: ", err)
	}

	output, err := cmd.Output()
	if err != nil {
		log.Errorln("Unable to verify yaml file:", string(output), err)
		return false
	}

	log.Infoln("Signature of yaml file is valid")
	return true
}

func setEnvVariablesForCommand(cmd *exec.Cmd, variables map[string]string) {
	getEnvVarName := func(key string) string {
		return fmt.Sprintf("RHC_WORKER_%s", strings.ToUpper(key))
	}
	for key, value := range variables {
		prefixedKey := getEnvVarName(key)
		envVarSetString := fmt.Sprintf("%s=%s", prefixedKey, value)
		cmd.Env = append(cmd.Env, envVarSetString)
		log.Infoln("Successfully set env variable", prefixedKey)
	}
}

// Parses given yaml data.
// If signature is valid then extracts the script to temporary file,
// sets env variables if present and then runs the script.
// Return stdout of executed script or error message if the signature wasn't valid.
func processSignedScript(incomingContent []byte) string {
	if len(incomingContent) == 0 {
		err := "Incoming Yaml content is empty"
		log.Errorln(err)
		return ""
	}

	// Verify signature
	signatureIsValid := verifyYamlFile(incomingContent)
	if !signatureIsValid {
		errorMsg := "Signature of yaml file is invalid"
		return errorMsg
	}

	// Parse the YAML data into array consisting of items of expected structure
	var signedYamlArray []signedYamlContent
	err := yaml.Unmarshal(incomingContent, &signedYamlArray)
	if err != nil {
		log.Errorln(err)
		return "Yaml couldn't be unmarshaled"
	}

	// We know/expect that the incoming data are array with ONLY one item
	yamlContent := signedYamlArray[0]

	// Write the file contents to the temporary disk
	log.Infof("Writing temporary script to %s", *config.TemporaryWorkerDirectory)
	scriptFileName := writeFileToTemporaryDir(
		[]byte(yamlContent.Vars.Content), *config.TemporaryWorkerDirectory)
	defer os.Remove(scriptFileName)

	log.Infoln("Processing script ...")

	// Execute script
	log.Infoln("Executing script...")
	cmd := exec.Command(yamlContent.Vars.Interpreter, scriptFileName) //nolint:gosec
	cmd.Env = os.Environ()

	variables := yamlContent.Vars.ContentVars
	variables["LOG_LEVEL"] = *config.ScriptLogLevel

	// Set env vars from yaml envelope
	setEnvVariablesForCommand(cmd, variables)

	// Set env vars from config
	if config.Env != nil {
		setEnvVariablesForCommand(cmd, *config.Env)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Errorln("Failed to execute script: ", err)
		if len(out) > 0 {
			log.Errorln(string(out))
		}
		return ""
	}

	log.Infoln("Script executed successfully.")
	return string(out)
}
