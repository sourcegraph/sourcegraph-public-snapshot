package cliutil

import (
	"context"
	"flag"
	"fmt"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Up(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	var (
		upFlagSet          = flag.NewFlagSet(fmt.Sprintf("%s up", commandName), flag.ExitOnError)
		upDatabaseNameFlag = upFlagSet.String("db", "all", `The target database instance. Supply "all" (the default) to migrate all databases.`)
		upTargetFlag       = upFlagSet.Int("target", 0, "Apply all migrations up to this target. Zero (the default) applies all migrations.")
	)

	execUp := func(ctx context.Context, args []string) error {
		if len(args) != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		if *upDatabaseNameFlag == "all" && *upTargetFlag != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply -db to migrate a specific database"))
			return flag.ErrHelp
		}

		var databaseNames []string
		if *upDatabaseNameFlag == "all" {
			databaseNames = append(databaseNames, schemas.SchemaNames...)
		} else {
			databaseNames = append(databaseNames, *upDatabaseNameFlag)
		}

		return run(ctx, runner.Options{
			Up:              true,
			TargetMigration: *upTargetFlag,
			SchemaNames:     databaseNames,
		})
	}

	return &ffcli.Command{
		Name:       "up",
		ShortUsage: fmt.Sprintf("%s up [-db=all] [-target=0]", commandName),
		ShortHelp:  "Run up migrations",
		FlagSet:    upFlagSet,
		Exec:       execUp,
		LongHelp:   ConstructLongHelp(),
	}
}
