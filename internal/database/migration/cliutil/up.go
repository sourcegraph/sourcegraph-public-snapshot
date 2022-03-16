package cliutil

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Up(commandName string, factory RunnerFactory, out *output.Output, development bool) *ffcli.Command {
	var (
		flagSet                  = flag.NewFlagSet(fmt.Sprintf("%s up", commandName), flag.ExitOnError)
		schemaNameFlag           = flagSet.String("db", "all", `The target schema(s) to modify. Comma-separated values are accepted. Supply "all" (the default) to migrate all schemas.`)
		unprivilegedOnlyFlag     = flagSet.Bool("unprivileged-only", false, `Do not apply privileged migrations.`)
		ignoreSingleDirtyLogFlag = flagSet.Bool("ignore-single-dirty-log", development, `Ignore a previously failed attempt if it will be immediately retried by this operation.`)
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
		sort.Strings(schemaNames)

		operations := []runner.MigrationOperation{}
		for _, schemaName := range schemaNames {
			operations = append(operations, runner.MigrationOperation{
				SchemaName: schemaName,
				Type:       runner.MigrationOperationTypeUpgrade,
			})
		}

		r, err := factory(ctx, schemaNames)
		if err != nil {
			return err
		}

		return r.Run(ctx, runner.Options{
			Operations:           operations,
			UnprivilegedOnly:     *unprivilegedOnlyFlag,
			IgnoreSingleDirtyLog: *ignoreSingleDirtyLogFlag,
		})
	}

	return &ffcli.Command{
		Name:       "up",
		ShortUsage: fmt.Sprintf("%s up [-db=<schema>]", commandName),
		ShortHelp:  "Apply all migrations",
		FlagSet:    flagSet,
		Exec:       exec,
		LongHelp:   ConstructLongHelp(),
	}
}
