package cliutil

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Up(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	var (
		flagSet        = flag.NewFlagSet(fmt.Sprintf("%s up", commandName), flag.ExitOnError)
		schemaNameFlag = flagSet.String("db", "all", `The target schema(s) to migrate. Comma-separated values are accepted. Supply "all" (the default) to migrate all schemas.`)
	)

	exec := func(ctx context.Context, args []string) error {
		if len(args) != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		schemaNames := strings.Split(*schemaNameFlag, ",")
		if len(schemaNames) == 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a schema via -db"))
			return flag.ErrHelp
		}
		if len(schemaNames) == 1 && schemaNames[0] == "all" {
			schemaNames = schemas.SchemaNames
		}

		operations := []runner.MigrationOperation{}
		for _, schemaName := range schemaNames {
			operations = append(operations, runner.MigrationOperation{
				SchemaName: schemaName,
				Type:       runner.MigrationOperationTypeUpgrade,
			})
		}

		return run(ctx, runner.Options{
			Operations: operations,
		})
	}

	return &ffcli.Command{
		Name:       "up",
		ShortUsage: fmt.Sprintf("%s up -db=<schema>", commandName),
		ShortHelp:  "Apply all migrations",
		FlagSet:    flagSet,
		Exec:       exec,
		LongHelp:   ConstructLongHelp(),
	}
}
