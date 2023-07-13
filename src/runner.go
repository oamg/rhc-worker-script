package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"git.sr.ht/~spc/go-log"
	"gopkg.in/yaml.v3"
)

// Received Yaml data has to match the expected yamlConfig structure
type yamlConfig struct {
	Vars struct {
		InsightsSignature        string            `yaml:"_insights_signature"`
		InsightsSignatureExclude string            `yaml:"_insights_signature_exclude"`
		Content                  string            `yaml:"content"`
		ContentVars              map[string]string `yaml:"content_vars"`
	} `yaml:"vars"`
}

// Verify that no one tampered with yaml file
func verifyYamlFile(yamlData []byte) bool {

	if shouldVerifyYaml != "1" {
		log.Warnln("WARNING: Playbook verification disabled.")
		return true
	}

	log.Infoln("Verifying yaml file...")
	// --payload here will be a no-op because no upload is performed when using the verifier
	//   but, it will allow us to update the egg!

	args := []string{
		"-m", "insights.client.apps.ansible.playbook_verifier",
		"--quiet", "--payload", "noop", "--content-type", "noop",
	}
	env := os.Environ()

	if shouldDoInsightsCoreGPGCheck == "0" {
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

// Parses given yaml data.
// If signature is valid then extracts the bash script to temporary file,
// sets env variables if present and then runs the script.
// Return stdout of executed script or error message if the signature wasn't valid.
func processSignedScript(yamlFileContet []byte) string {
	signatureIsValid := verifyYamlFile(yamlFileContet)
	if !signatureIsValid {
		errorMsg := "Signature of yaml file is invalid"
		log.Errorln(errorMsg)
		return errorMsg
	}
	log.Infoln("Signature of yaml file is valid")

	// Parse the YAML data into the yamlConfig struct
	var yamlContent yamlConfig
	err := yaml.Unmarshal(yamlFileContet, &yamlContent)
	if err != nil {
		log.Errorln(err)
	}

	// Set env variables
	getEnvVarName := func(key string) string {
		return fmt.Sprintf("RHC_WORKER_%s", strings.ToUpper(key))
	}
	for key, value := range yamlContent.Vars.ContentVars {
		prefixedKey := getEnvVarName(key)
		err := os.Setenv(prefixedKey, value)
		if err != nil {
			log.Errorln(err)
		} else {
			log.Infoln("Successfully set env variable", prefixedKey, "=", value)
		}
	}
	defer func() {
		for key := range yamlContent.Vars.ContentVars {
			os.Unsetenv(getEnvVarName(key))
		}
	}()

	// NOTE: just debug to see the values
	log.Debugln("Insights Signature:", yamlContent.Vars.InsightsSignature)
	log.Debugln("Insights Signature Exclude:", yamlContent.Vars.InsightsSignatureExclude)
	log.Debugln("Script:", yamlContent.Vars.Content)
	log.Debugln("Vars:")
	for key, value := range yamlContent.Vars.ContentVars {
		log.Debugln(" ", key, ":", value)
	}

	// Write the file contents to the temporary disk
	log.Infoln("Writing temporary bash script")
	scriptFileName := writeFileToTemporaryDir([]byte(yamlContent.Vars.Content), temporaryWorkerDirectory)
	defer os.Remove(scriptFileName)

	// Execute the script
	log.Infoln("Executing bash script")

	out, err := exec.Command("/bin/sh", scriptFileName).Output()
	if err != nil {
		log.Errorln("Failed to execute script: ", err)
		return ""
	}

	log.Infoln("Bash script executed successfully")
	return string(out)
}
