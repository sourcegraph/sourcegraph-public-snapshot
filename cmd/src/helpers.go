package main

import (
	"os"
	"os/exec"
)

// verboseCmdOutput pipes a command's stdout/stderr to stderr if the `-v` (verbose) flag is enabled.
func verboseCmdOutput(cmd *exec.Cmd) {
	if *verbose {
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
	}
}
