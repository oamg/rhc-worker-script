package main

import (
	"os/exec"

	"git.sr.ht/~spc/go-log"
)

func executeScript(fileName string) {
	cmd := exec.Command("/bin/sh", fileName)
	log.Debug("Cmd: ", cmd)
	if err := cmd.Run(); err != nil {
		log.Errorln("Failed to execute script: ", err)
	}

	log.Infoln("Bash script executed successfully.")
}
