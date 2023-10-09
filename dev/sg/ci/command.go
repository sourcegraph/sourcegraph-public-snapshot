package ci

import (
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/category"
	"github.com/urfave/cli/v2"
)

var (
	ciBranchFlag = cli.StringFlag{
		Name:    "branch",
		Aliases: []string{"b"},
		Usage:   "Branch `name` of build to target (defaults to current branch)",
	}
	ciBuildFlag = cli.StringFlag{
		Name:    "build",
		Aliases: []string{"n"}, // 'n' for number, because 'b' is taken
		Usage:   "Override branch detection with a specific build `number`",
	}
	ciCommitFlag = cli.StringFlag{
		Name:    "commit",
		Aliases: []string{"c"},
		Usage:   "Override branch detection with the latest build for `commit`",
	}
	ciPipelineFlag = cli.StringFlag{
		Name:    "pipeline",
		Aliases: []string{"p"},
		EnvVars: []string{"SG_CI_PIPELINE"},
		Usage:   "Select a custom Buildkite `pipeline` in the Sourcegraph org",
		Value:   "sourcegraph",
	}
)

// ciTargetFlags register the following flags on all commands that can target different builds.
var ciTargetFlags = []cli.Flag{
	&ciBranchFlag,
	&ciBuildFlag,
	&ciCommitFlag,
	&ciPipelineFlag,
}

// Command is a top level command that provides a variety of CI subcommands.
var Command = &cli.Command{
	Name:        "ci",
	Usage:       "Interact with Sourcegraph's Buildkite continuous integration pipelines",
	Description: `Note that Sourcegraph's CI pipelines are under our enterprise license: https://github.com/sourcegraph/sourcegraph/blob/main/LICENSE.enterprise`,
	UsageText: `
# Preview what a CI run for your current changes will look like
sg ci preview

# Check on the status of your changes on the current branch in the Buildkite pipeline
sg ci status
# Check on the status of a specific branch instead
sg ci status --branch my-branch
# Block until the build has completed (it will send a system notification)
sg ci status --wait
# Get status for a specific build number
sg ci status --build 123456

# Pull logs of failed jobs to stdout
sg ci logs
# Push logs of most recent main failure to local Loki for analysis
# You can spin up a Loki instance with 'sg run loki grafana'
sg ci logs --branch main --out http://127.0.0.1:3100
# Get the logs for a specific build number, useful when debugging
sg ci logs --build 123456

# Manually trigger a build on the CI with the current branch
sg ci build
# Manually trigger a build on the CI on the current branch, but with a specific commit
sg ci build --commit my-commit
# Manually trigger a main-dry-run build of the HEAD commit on the current branch
sg ci build main-dry-run
sg ci build --force main-dry-run
# Manually trigger a main-dry-run build of a specified commit on the current ranch
sg ci build --force --commit my-commit main-dry-run
# View the available special build types
sg ci build --help
`,
	Category: category.Dev,
	Subcommands: []*cli.Command{
		previewCommand,
		bazelCommand,
		statusCommand,
		buildCommand,
		logsCommand,
		docsCommand,
		openCommand,
	},
}
