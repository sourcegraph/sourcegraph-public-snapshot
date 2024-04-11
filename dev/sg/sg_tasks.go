package main

import (
	"os"
	"os/exec"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/urfave/cli/v2"
)

var tasksCommand = &cli.Command{
	Name:        "tasks",
	Aliases:     []string{"t"},
	Usage:       "Run commands configured in external manifests.",
	Description: "Run commands configured in external manifests. See 'tasks' section of sg.config.yaml.",
	Category:    category.Util,
	Subcommands: tasksSubcommands(),
}

// TODO this is a placeholder/spike - change to use external manifests, as
// described in https://github.com/sourcegraph/devx-support/issues/801
func tasksSubcommands() []*cli.Command {
	cfg, err := preFlagConfig()
	if err != nil {
		// If there was an error loading the pre-flag config, we can't configure
		// tasks. It is possible for there to be errors loading the pre-flag
		// config when the config itself can be loaded after command
		// initialization - see comment on preFlagConfig(). Rather than break
		// the whole CLI by exitting in this edge case, we just don't load the
		// tasks.
		return nil
	}

	var cmds []*cli.Command
	for name, task := range cfg.Tasks {
		cmd := &cli.Command{
			Name:  name,
			Usage: task.Usage,
			Action: func(*cli.Context) error {
				// Using a shell seems appropriate rather than splitting the
				// command by whitespace. This feature is intended to replace
				// shell aliases/functions, so by using a shell we get PATH
				// resolution and expected quoting behavior for free.
				//
				// Arguably we should use $SHELL but /bin/sh is probably fine
				// for most cases.
				shellCmd := exec.Command("sh", "-c", task.Command)
				shellCmd.Stdout = os.Stdout
				shellCmd.Stderr = os.Stderr
				return shellCmd.Run()
			},
		}
		cmds = append(cmds, cmd)
	}
	return cmds
}
