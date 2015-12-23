package worker

import (
	"log"

	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	g, err := cli.CLI.AddCommand("internal-build",
		"(internal) build worker commands",
		"The internal-build subcommands are internal commands for the build worker.",
		&internalBuildsCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.AddCommand("prep",
		"(internal) prepare a build dir",
		`
Prepares a build directory at BUILD-DIR for the repository build with the
specified build ID. This involves cloning the repository and fetching cached
build data (if any).

This is an internal sub-command that is not typically invoked directly.
`,
		&prepBuildCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.AddCommand("run",
		"invoke 'build prep' and 'build do'",
		`
Runs a build (specified by BUILD-ID) in BUILD-DIR. This command simply invokes
"sourcegraph 'build prep'" and "sourcegraph 'build do'".

This is an internal sub-command that is not typically invoked directly.`,
		&runBuildCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.AddCommand("do",
		"(internal) run a build",
		`Runs the repository build with the specified build ID in the current directory.

This is an internal sub-command that is not typically invoked directly. If you
invoke it directly, it will create build tasks in the DB but will not update the
build itself, so the build will still appear uncompleted. The worker (not this
subcommand) is responsible for updating the build.`,
		&doBuildCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type internalBuildsCmd struct{}

func (c *internalBuildsCmd) Execute(args []string) error { return nil }
