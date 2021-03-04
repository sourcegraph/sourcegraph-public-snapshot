package graph

import (
	"os/exec"
	"strings"
)

// runGo invokes a go command on the host.
func runGo(commands ...string) (string, error) {
	cmd := exec.Command("go", commands...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(out)), nil
}
