package cliutil

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Undo(commandName string, factory RunnerFactory, out *output.Output, development bool) *ffcli.Command {
	var (
		flagSet                  = flag.NewFlagSet(fmt.Sprintf("%s undo", commandName), flag.ExitOnError)
		schemaNameFlag           = flagSet.String("db", "", `The target schema to modify.`)
		ignoreSingleDirtyLogFlag = flagSet.Bool("ignore-single-dirty-log", development, `Ignore a previously failed attempt if it will be immediately retried by this operation.`)
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

		r, err := factory(ctx, []string{*schemaNameFlag})
		if err != nil {
			return err
		}

		return r.Run(ctx, runner.Options{
			Operations: []runner.MigrationOperation{
				{
					SchemaName: *schemaNameFlag,
					Type:       runner.MigrationOperationTypeRevert,
				},
			},
			IgnoreSingleDirtyLog: *ignoreSingleDirtyLogFlag,
		})
	}

	return &ffcli.Command{
		Name:       "undo",
		ShortUsage: fmt.Sprintf("%s undo -db=<schema>", commandName),
		ShortHelp:  `Revert the last migration applied - useful in local development`,
		FlagSet:    flagSet,
		Exec:       exec,
		LongHelp:   ConstructLongHelp(),
	}
}
