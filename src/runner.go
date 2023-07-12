package main

import (
	"os/exec"

	"git.sr.ht/~spc/go-log"
)

// Executes bash script and returns its standard output
func executeScript(fileName string) string {
	out, err := exec.Command("/bin/sh", fileName).Output()
	if err != nil {
		log.Errorln("Failed to execute script: ", err)
		return ""
	}

	log.Infoln("Bash script executed successfully.")
	return string(out)
}
