//go:build !shell

package command

import "github.com/sourcegraph/log"

// NewRunner creates a new runner with the given options.
func NewRunner(dir string, logger Logger, options Options, operations *Operations) Runner {
	if !options.FirecrackerOptions.Enabled {
		return &dockerRunner{
			dir:       dir,
			logger:    log.Scoped("docker-runner", ""),
			cmdLogger: logger,
			options:   options,
		}
	}

	return &firecrackerRunner{
		name:            options.ExecutorName,
		workspaceDevice: dir,
		logger:          logger,
		options:         options,
		operations:      operations,
	}
}
