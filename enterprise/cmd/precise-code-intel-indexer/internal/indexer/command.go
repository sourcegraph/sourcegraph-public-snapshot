package indexer

import (
	"fmt"
	"os/exec"

	"github.com/pkg/errors"
)

func command(dir, command string, args ...string) error {
	indexCmd := exec.Command(command, args...)
	indexCmd.Dir = dir

	if output, err := indexCmd.CombinedOutput(); err != nil {
		return errors.Wrap(err, fmt.Sprintf("command failed: %s\n", output))
	}

	return nil
}
