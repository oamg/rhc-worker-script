package main

import (
	"os/exec"

	"git.sr.ht/~spc/go-log"
)

// TODO(r0x0d): Add docstring
func executeScript(fileName string) {
	log.Info("Temporary bash script created at: ", fileName)
	cmd := exec.Command("/bin/sh", fileName)
	log.Debug("Cmd: ", cmd)
	if err := cmd.Run(); err != nil {
		log.Errorln("error: ", err)
	}
}
