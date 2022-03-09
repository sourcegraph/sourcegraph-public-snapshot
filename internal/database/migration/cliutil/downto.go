package cliutil

import (
	"context"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func DownTo(commandName string, factory RunnerFactory, out *output.Output, development bool) *ffcli.Command {
	var (
		flagSet                  = flag.NewFlagSet(fmt.Sprintf("%s downto", commandName), flag.ExitOnError)
		schemaNameFlag           = flagSet.String("db", "", `The target schema to modify.`)
		unprivilegedOnlyFlag     = flagSet.Bool("unprivileged-only", false, `Do not apply privileged migrations.`)
		ignoreSingleDirtyLogFlag = flagSet.Bool("ignore-single-dirty-log", development, `Ignore a previously failed attempt if it will be immediately retried by this operation.`)
		targetsFlag              = flagSet.String("target", "", "Revert all children of the given target. Comma-separated values are accepted.")
	)

	exec := func(ctx context.Context, args []string) error {
		if len(args) != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		if *schemaNameFlag == "" {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a schema via -db"))
			return flag.ErrHelp
		}

		targets := strings.Split(*targetsFlag, ",")
		if len(targets) == 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a migration target via -target"))
			return flag.ErrHelp
		}

		versions := make([]int, 0, len(targets))
		for _, target := range targets {
			version, err := strconv.Atoi(target)
			if err != nil {
				return err
			}

			versions = append(versions, version)
		}

		r, err := factory(ctx, []string{*schemaNameFlag})
		if err != nil {
			return err
		}

		return r.Run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName:     *schemaNameFlag,
					Type:           runner.MigrationOperationTypeTargetedDown,
					TargetVersions: versions,
				},
			},
			UnprivilegedOnly:     *unprivilegedOnlyFlag,
			IgnoreSingleDirtyLog: *ignoreSingleDirtyLogFlag,
		})
	}

	return &ffcli.Command{
		Name:       "downto",
		ShortUsage: fmt.Sprintf("%s downto -db=<schema> -target=<target>,<target>,...", commandName),
		ShortHelp:  `Revert any applied migrations that are children of the given targets - this effectively "resets" the schmea to the target version`,
		FlagSet:    flagSet,
		Exec:       exec,
		LongHelp:   ConstructLongHelp(),
	}
}
