package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git.sr.ht/~spc/go-log"
	"gopkg.in/yaml.v3"
)

// Received Yaml data has to match the expected signedYamlFile structure
type signedYamlFile struct {
	Vars struct {
		InsightsSignature        string            `yaml:"_insights_signature"`
		InsightsSignatureExclude string            `yaml:"_insights_signature_exclude"`
		Content                  string            `yaml:"content"`
		ContentVars              map[string]string `yaml:"content_vars"`
	} `yaml:"vars"`
}

// Verify that no one tampered with yaml file
func verifyYamlFile(yamlData []byte) bool {

	if !*config.VerifyYAML {
		log.Warnln("WARNING: Playbook verification disabled.")
		return true
	}

	// --payload here will be a no-op because no upload is performed when using the verifier
	//   but, it will allow us to update the egg!

	args := []string{
		"-m", "insights.client.apps.ansible.playbook_verifier",
		"--quiet", "--payload", "noop", "--content-type", "noop",
	}
	env := os.Environ()

	if !*config.InsightsCoreGPGCheck {
		args = append(args, "--no-gpg")
		env = append(env, "BYPASS_GPG=True")
	}

	cmd := exec.Command("insights-client", args...)
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
	stdin.Close()

	output, err := cmd.Output()
	if err != nil {
		log.Errorln("ERROR: Unable to verify yaml file:", string(output), err)
		return false
	}
	return true
}

func setEnvVariablesForCommand(cmd *exec.Cmd, variables map[string]string) {
	cmd.Env = os.Environ()
	getEnvVarName := func(key string) string {
		return fmt.Sprintf("RHC_WORKER_%s", strings.ToUpper(key))
	}
	for key, value := range variables {
		prefixedKey := getEnvVarName(key)
		envVarSetString := fmt.Sprintf("%s=%s", prefixedKey, value)
		cmd.Env = append(cmd.Env, envVarSetString)
		log.Infoln("Successfully set env variable ", prefixedKey)
	}
}

// Parses given yaml data.
// If signature is valid then extracts the bash script to temporary file,
// sets env variables if present and then runs the script.
// Return stdout of executed script or error message if the signature wasn't valid.
func processSignedScript(incomingContent []byte) string {
	// Verify signature
	log.Infoln("Verifying signature ...")
	signatureIsValid := verifyYamlFile(incomingContent)
	if !signatureIsValid {
		errorMsg := "Signature of yaml file is invalid"
		return errorMsg
	}
	log.Infoln("Signature of yaml file is valid")

	// Parse the YAML data into the yamlConfig struct
	var yamlContent signedYamlFile
	err := yaml.Unmarshal(incomingContent, &yamlContent)
	if err != nil {
		log.Errorln(err)
		return "Yaml couldn't be unmarshaled"
	}

	// Write the file contents to the temporary disk
	log.Infoln("Writing temporary bash script")
	scriptFileName := writeFileToTemporaryDir(
		[]byte(yamlContent.Vars.Content), *config.TemporaryWorkerDirectory)
	defer os.Remove(scriptFileName)

	log.Infoln("Processing bash script ...")

	// Execute script
	log.Infoln("Executing bash script...")
	cmd := exec.Command("/bin/sh", scriptFileName)
	setEnvVariablesForCommand(cmd, yamlContent.Vars.ContentVars)

	out, err := cmd.Output()
	if err != nil {
		log.Errorln("Failed to execute script: ", err)
		return ""
	}

	log.Infoln("Bash script executed successfully")
	return string(out)
}
