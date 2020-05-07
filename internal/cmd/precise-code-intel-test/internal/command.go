package internal

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

// runCommand runs the command in the given working dir and returns an error.
func runCommand(dir, command string, args ...string) error {
	_, err := runCommandOutput(dir, command, args...)
	return err
}

// runCommand runs the commandin the given working dir and returns its output and an error.
func runCommandOutput(dir, command string, args ...string) ([]byte, error) {
	cmd := exec.Command(command, args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("error running '%s':\n%s\n", strings.Join(append([]string{command}, args...), " "), output))
	}

	return output, nil
}
