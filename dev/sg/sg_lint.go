package main

import (
	"flag"
	"strings"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/repo"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
	"github.com/sourcegraph/sourcegraph/dev/sg/linters"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var generateAnnotations = &cli.BoolFlag{
	Name:  "annotations",
	Usage: "Write helpful output to ./annotations directory",
}

var lintFix = &cli.BoolFlag{
	Name:    "fix",
	Aliases: []string{"f"},
	Usage:   "Try to fix any lint issues",
}

var lintFailFast = &cli.BoolFlag{
	Name:    "fail-fast",
	Aliases: []string{"ff"},
	Usage:   "Exit immediately if an issue is encountered (not available with '-fix')",
	Value:   true,
}

var lintCommand = &cli.Command{
	Name:        "lint",
	ArgsUsage:   "[targets...]",
	Usage:       "Run all or specified linters on the codebase",
	Description: `To run all checks, don't provide an argument. You can also provide multiple arguments to run linters for multiple targets.`,
	UsageText: `
# Run all possible checks
sg lint

# Run only go related checks
sg lint go

# Run only shell related checks
sg lint shell

# Run only client related checks
sg lint client

# List all available check groups
sg lint --help
`,
	Category: CategoryDev,
	Flags: []cli.Flag{
		generateAnnotations,
		lintFix,
		lintFailFast,
	},
	Before: func(cmd *cli.Context) error {
		// If more than 1 target is requested, hijack subcommands by setting it to nil
		// so that the main lint command can handle it the run.
		if cmd.Args().Len() > 1 {
			cmd.App.Commands = nil
		}
		return nil
	},
	Action: func(cmd *cli.Context) error {
		var lintTargets []linters.Target
		targets := cmd.Args().Slice()

		if len(targets) == 0 {
			// If no args provided, run all
			lintTargets = linters.Targets
			for _, t := range lintTargets {
				targets = append(targets, t.Name)
			}
		} else {
			// Otherwise run requested set
			allLintTargetsMap := make(map[string]linters.Target, len(linters.Targets))
			for _, c := range linters.Targets {
				allLintTargetsMap[c.Name] = c
			}
			for _, t := range targets {
				target, ok := allLintTargetsMap[t]
				if !ok {
					std.Out.WriteFailuref("unrecognized target %q provided", t)
					return flag.ErrHelp
				}
				lintTargets = append(lintTargets, target)
			}
		}

		repoState, err := repo.GetState(cmd.Context)
		if err != nil {
			return errors.Wrap(err, "repo.GetState")
		}

		runner := linters.NewRunner(std.Out, generateAnnotations.Get(cmd), lintTargets...)
		if cmd.Bool("fix") {
			std.Out.WriteNoticef("Fixing checks from targets: %s", strings.Join(targets, ", "))
			return runner.Fix(cmd.Context, repoState)
		}
		runner.FailFast = lintFailFast.Get(cmd)
		std.Out.WriteNoticef("Running checks from targets: %s", strings.Join(targets, ", "))
		return runner.Check(cmd.Context, repoState)
	},
	Subcommands: lintTargets(linters.Targets).Commands(),
}

type lintTargets []linters.Target

// Commands converts all lint targets to CLI commands
func (lt lintTargets) Commands() (cmds []*cli.Command) {
	for _, target := range lt {
		target := target // local reference
		cmds = append(cmds, &cli.Command{
			Name:  target.Name,
			Usage: target.Description,
			Action: func(cmd *cli.Context) error {
				if cmd.NArg() > 0 {
					std.Out.WriteFailuref("unrecognized argument %q provided", cmd.Args().First())
					return flag.ErrHelp
				}

				repoState, err := repo.GetState(cmd.Context)
				if err != nil {
					return errors.Wrap(err, "repo.GetState")
				}

				runner := linters.NewRunner(std.Out, generateAnnotations.Get(cmd), target)
				if lintFix.Get(cmd) {
					std.Out.WriteNoticef("Fixing checks from target: %s", target.Name)
					return runner.Fix(cmd.Context, repoState)
				}
				runner.FailFast = lintFailFast.Get(cmd)
				std.Out.WriteNoticef("Running checks from target: %s", target.Name)
				return runner.Check(cmd.Context, repoState)
			},
			// Completions to chain multiple commands
			BashComplete: completeOptions(func() (options []string) {
				for _, c := range lt {
					options = append(options, c.Name)
				}
				return options
			}),
		})
	}
	return cmds
}
