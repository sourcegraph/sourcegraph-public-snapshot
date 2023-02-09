package runtime

import (
	"os/exec"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/internal/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func validateDockerRuntime() ([]string, error) {
	var notFoundTools []string
	for tool := range config.RequiredCLITools {
		if found, err := existsPath(tool); err != nil {
			return notFoundTools, err
		} else if !found {
			notFoundTools = append(notFoundTools, tool)
		}
	}
	return notFoundTools, nil
}

func existsPath(name string) (bool, error) {
	if _, err := exec.LookPath(name); err != nil {
		if errors.Is(err, exec.ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
