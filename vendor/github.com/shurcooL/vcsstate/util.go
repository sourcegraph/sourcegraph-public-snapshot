package vcsstate

import (
	"bytes"
	"os/exec"
)

// dividedOutput runs the command and returns its standard output and standard error.
func dividedOutput(cmd *exec.Cmd) (stdout []byte, stderr []byte, err error) {
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	err = cmd.Run()
	return outb.Bytes(), errb.Bytes(), err
}
