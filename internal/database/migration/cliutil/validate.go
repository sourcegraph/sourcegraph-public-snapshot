package cliutil

import (
	"context"
	"flag"
	"fmt"
	"sort"
	"strings"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Validate(commandName string, factory RunnerFactory, out *output.Output) *ffcli.Command {
	var (
		flagSet        = flag.NewFlagSet(fmt.Sprintf("%s validate", commandName), flag.ExitOnError)
		schemaNameFlag = flagSet.String("db", "all", `The target schema(s) to validate. Comma-separated values are accepted. Supply "all" (the default) to validate all schemas.`)
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

		r, err := factory(ctx, schemaNames)
		if err != nil {
			return err
		}

		return r.Validate(ctx, schemaNames...)
	}

	return &ffcli.Command{
		Name:       "validate",
		ShortUsage: fmt.Sprintf("%s validate [-db=<schema>]", commandName),
		ShortHelp:  "Validate the current schema",
		FlagSet:    flagSet,
		Exec:       exec,
		LongHelp:   ConstructLongHelp(),
	}
}
