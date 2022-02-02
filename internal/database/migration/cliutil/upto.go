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

func UpTo(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	var (
		flagSet        = flag.NewFlagSet(fmt.Sprintf("%s upto", commandName), flag.ExitOnError)
		schemaNameFlag = flagSet.String("db", "", `The target schema to migrate.`)
		targetFlag     = flagSet.String("target", "", "The migration to apply.")
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
					Type:          runner.MigrationOperationTypeTargetedUp,
					TargetVersion: version,
				},
			},
		})
	}

	return &ffcli.Command{
		Name:       "upto",
		ShortUsage: fmt.Sprintf("%s upto -db=<schema> -target=<target>,<target>,...", commandName),
		ShortHelp:  "Ensure a given migration has been applied - may apply dependency migrations",
		FlagSet:    flagSet,
		Exec:       exec,
		LongHelp:   ConstructLongHelp(),
	}
}
