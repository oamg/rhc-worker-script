package main

import (
	"os/exec"

	"git.sr.ht/~spc/go-log"
)

func executeScript(fileName string) string {
	out, err := exec.Command("/bin/sh", fileName).Output()
	if err != nil {
		log.Errorln("Failed to execute script: ", err)
	}

	log.Infoln("Bash script executed successfully.")
	return string(out)
}
