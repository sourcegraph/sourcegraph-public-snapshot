package cliutil

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func DownTo(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	var (
		flagSet        = flag.NewFlagSet(fmt.Sprintf("%s downto", commandName), flag.ExitOnError)
		schemaNameFlag = flagSet.String("db", "", `The target schema to migrate.`)
		targetFlag     = flagSet.String("target", "", "Revert all children of the given target.")
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

		if *targetFlag == "" {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a migration target via -target"))
			return flag.ErrHelp
		}

		version, err := strconv.Atoi(*targetFlag)
		if err != nil {
			return err
		}

		return run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName:    *schemaNameFlag,
					Type:          runner.MigrationOperationTypeTargetedDown,
					TargetVersion: version,
				},
			},
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
