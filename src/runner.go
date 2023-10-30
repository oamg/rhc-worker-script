package main

import (
	"bufio"
	"fmt"
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
func runCommandWithOutput(cmd *exec.Cmd, outputCh chan []byte, doneCh chan bool) {
	cmdOutput, err := cmd.StdoutPipe()
	if err != nil {
		log.Errorln("Error: ", err)
		doneCh <- true
		return
	}

	dataReadCh := make(chan bool)
	defer close(dataReadCh)

	go func() {
		// TODO: not good for a buffer to have constant size ...
		// Need to have some implementation of reseting the full byte slice
		scanner := bufio.NewScanner(cmdOutput)
		bufferSize := 1024
		scanner.Buffer(make([]byte, bufferSize), bufferSize/4)

		for {
			for scanner.Scan() {
				line := scanner.Text()
				outputCh <- []byte(line + "\n")
			}

			if err := scanner.Err(); err == nil {
				// Scanner reached EOF
				break
			} else {
				// TODO: error bufio.Scanner: token too long but it looks like we are not loosing data this way
				log.Infoln(err)
				scanner = bufio.NewScanner(cmdOutput)
				scanner.Buffer(make([]byte, bufferSize), bufferSize/4)
			}
		}
		dataReadCh <- true

		//////////

		// NOTE: below code also works but it's dependent on default go buffer
		// in our case it just means that the stdout is reported at the end

		// reader := bufio.NewReader(cmdOutput)
		// readNBufferBytes := 1024

		// for {
		// 	data, err := reader.Peek(readNBufferBytes)
		// 	switch {
		// 	case errors.Is(err, io.EOF):
		// 		log.Infoln("Read ended with EOF")
		// 		outputCh <- data
		// 		dataReadCh <- true
		// 		return
		// 	case err == nil || errors.Is(err, io.ErrShortBuffer):
		// 		log.Infoln("Read n bytes", err)
		// 		if len(data) != 0 {
		// 			outputCh <- data
		// 			_, err := reader.Discard(len(data))
		// 			if err != nil {
		// 				// TODO: what should I do if I want to move the reader
		// 				// If I do nothing it can only cause it to be read twice
		// 				log.Errorln("Discard failed", err)
		// 			}
		// 		}
		// 	}
		// }
	}()

	if err := cmd.Start(); err != nil {
		log.Errorln("Error: ", err)
		doneCh <- true
		return
	}

	// NOTE: need to block here before goroutine finishes so wait doesn't close the stdout pipe
	log.Infoln("Waiting to collect all stdout from running command")
	<-dataReadCh

	if err := cmd.Wait(); err != nil {
		log.Errorln("Failed to execute script: ", err)
	}

	doneCh <- true // Signal that the command has finished
}

// Executes command and reports status back to dispatcher
func executeCommandWithProgress(command string, interpreter string, variables map[string]string) string {
	log.Infoln("Executing script...")

	cmd := exec.Command(interpreter, command)
	setEnvVariablesForCommand(cmd, variables)

	var bufferedOutput []byte
	outputCh := make(chan []byte)
	defer close(outputCh)
	doneCh := make(chan bool)
	defer close(doneCh)

	go runCommandWithOutput(cmd, outputCh, doneCh)

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case output := <-outputCh:
			bufferedOutput = append(bufferedOutput, output...)
		case <-ticker.C:
			// NOTE: If just message without output is also okay we could send just still running
			log.Infoln("Still running ...")
			log.Infoln(string(bufferedOutput))
		case <-doneCh:
			// Execution is done
			log.Infoln("Execution done ...")
			return string(bufferedOutput)
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
