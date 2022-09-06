package cliutil

import (
	"context"
	"fmt"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func UpTo(commandName string, factory RunnerFactory, outFactory OutputFactory, development bool) *cli.Command {
	schemaNameFlag := &cli.StringFlag{
		Name:     "db",
		Usage:    "The target `schema` to modify.",
		Required: true,
	}
	targetFlag := &cli.StringSliceFlag{
		Name:     "target",
		Usage:    "The `migration` to apply. Comma-separated values are accepted.",
		Required: true,
	}
	unprivilegedOnlyFlag := &cli.BoolFlag{
		Name:  "unprivileged-only",
		Usage: "Refuse to apply privileged migrations.",
		Value: false,
	}
	noopPrivilegedFlag := &cli.BoolFlag{
		Name:  "noop-privileged",
		Usage: "Skip application of privileged migrations, but record that they have been applied. This assumes the user has already applied the required privileged migrations with elevated permissions.",
		Value: false,
	}
	privilegedHashFlag := &cli.StringFlag{
		Name:  "privileged-hash",
		Usage: "Running -noop-privileged without this value will supply a value that will unlock migration application for the current upgrade operation. Future (distinct) upgrade operations will require a unique hash.",
		Value: "",
	}
	ignoreSingleDirtyLogFlag := &cli.BoolFlag{
		Name:  "ignore-single-dirty-log",
		Usage: "Ignore a single previously failed attempt if it will be immediately retried by this operation.",
		Value: development,
	}
	ignoreSinglePendingLogFlag := &cli.BoolFlag{
		Name:  "ignore-single-pending-log",
		Usage: "Ignore a single pending migration attempt if it will be immediately retried by this operation.",
		Value: development,
	}

	makeOptions := func(cmd *cli.Context, out *output.Output, versions []int) (runner.Options, error) {
		privilegedMode, err := getPivilegedModeFromFlags(cmd, out, unprivilegedOnlyFlag, noopPrivilegedFlag)
		if err != nil {
			return runner.Options{}, err
		}

		return runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName:     schemaNameFlag.Get(cmd),
					Type:           runner.MigrationOperationTypeTargetedUp,
					TargetVersions: versions,
				},
			},
			PrivilegedMode:         privilegedMode,
			PrivilegedHash:         privilegedHashFlag.Get(cmd),
			IgnoreSingleDirtyLog:   ignoreSingleDirtyLogFlag.Get(cmd),
			IgnoreSinglePendingLog: ignoreSinglePendingLogFlag.Get(cmd),
		}, nil
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		versions, err := parseTargets(targetFlag.Get(cmd))
		if err != nil {
			return err
		}
		if len(versions) == 0 {
			return flagHelp(out, "supply a target via -target")
		}

		r, err := setupRunner(ctx, factory, schemaNameFlag.Get(cmd))
		if err != nil {
			return err
		}

		options, err := makeOptions(cmd, out, versions)
		if err != nil {
			return err
		}

		return r.Run(ctx, options)
	})

	return &cli.Command{
		Name:        "upto",
		UsageText:   fmt.Sprintf("%s upto -db=<schema> -target=<target>,<target>,...", commandName),
		Usage:       "Ensure a given migration has been applied - may apply dependency migrations",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNameFlag,
			targetFlag,
			unprivilegedOnlyFlag,
			noopPrivilegedFlag,
			privilegedHashFlag,
			ignoreSingleDirtyLogFlag,
			ignoreSinglePendingLogFlag,
		},
	}
}
