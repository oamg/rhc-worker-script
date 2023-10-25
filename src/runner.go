package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

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

// Runs given command and sends stdout to given channel. doneCh used to signal that execution ended.
func runCommandWithOutput(cmd *exec.Cmd, outputCh chan string, doneCh chan bool) {
	cmdOutput, err := cmd.StdoutPipe()
	if err != nil {
		outputCh <- fmt.Sprintf("Error: %v", err)
		doneCh <- true
		return
	}

	go func() {
		// NOTE: This could be set by config or by env variable in received yaml
		var bufferSize = 512
		buf := make([]byte, bufferSize)
		for {
			log.Infoln("Writing to buffer ...")
			n, err := cmdOutput.Read(buf)
			log.Infoln("Buffer full ...")

			if err != nil {
				if err == io.EOF {
					log.Infoln("Command stdout closed")
					return
				}
				log.Infoln("Error after buffer closed")
				return
			}
			log.Infoln("Writing data to output channel ...")
			outputCh <- string(buf[:n])
			buf = make([]byte, bufferSize)
		}
	}()

	if err := cmd.Start(); err != nil {
		outputCh <- fmt.Sprintf("Error: %v", err)
		doneCh <- true
		return
	}

	if err := cmd.Wait(); err != nil {
		log.Errorln("Failed to execute script: ", err)
		outputCh <- fmt.Sprintf("Error: %v", err)
	}

	doneCh <- true // Signal that the command has finished
}

// Executes command and reports status back to dispatcher
func executeCommandWithProgress(command string, interpreter string, variables map[string]string) string {
	log.Infoln("Executing script...")

	cmd := exec.Command(interpreter, command)
	setEnvVariablesForCommand(cmd, variables)

	var bufferedOutput string
	outputCh := make(chan string)
	defer close(outputCh)
	doneCh := make(chan bool)
	defer close(doneCh)

	go runCommandWithOutput(cmd, outputCh, doneCh)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case output := <-outputCh:
			// TODO: this has to be sent to dispatcher back to report to UI
			// the idea is to send partial output if buffer with given size sent the output to channel
			log.Info(output)

			// Append partial to all output
			bufferedOutput += output
		case <-ticker.C:
			// NOTE: If just message without output is also okay we could send just still running
			log.Infoln("Still running ...")
		case <-doneCh:
			// Execution is done
			log.Infoln("Execution done ...")
			return bufferedOutput
		}
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

	// Execute script
	log.Infoln("Processing script ...")
	out := executeCommandWithProgress(
		scriptFileName, yamlContent.Vars.Interpreter, yamlContent.Vars.ContentVars)
	return out
}
