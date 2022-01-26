package cliutil

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Down(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	var (
		downFlagSet          = flag.NewFlagSet(fmt.Sprintf("%s down", commandName), flag.ExitOnError)
		downDatabaseNameFlag = downFlagSet.String("db", "", "The target database instance.")
		downTargetFlag       = downFlagSet.Int("target", 0, "Reset all migrations defined after this target. Zero (the default) reverts the latest migration.")
	)

	execDown := func(ctx context.Context, args []string) error {
		if len(args) != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		if *downDatabaseNameFlag == "" {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply -db to migrate a specific database"))
			return flag.ErrHelp
		}

		return run(ctx, runner.Options{
			Up:              false,
			TargetMigration: *downTargetFlag,
			SchemaNames:     []string{*downDatabaseNameFlag},
		})
	}

	return &ffcli.Command{
		Name:       "down",
		ShortUsage: fmt.Sprintf("%s down -db=... [-target=0]", commandName),
		ShortHelp:  "Run down migrations",
		FlagSet:    downFlagSet,
		Exec:       execDown,
		LongHelp:   ConstructLongHelp(),
	}
}
