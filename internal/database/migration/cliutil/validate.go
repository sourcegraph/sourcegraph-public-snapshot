package cliutil

import (
	"flag"

	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func Validate(commandName string, factory RunnerFactory, outFactory func() *output.Output) *cli.Command {
	flags := []cli.Flag{
		&cli.StringSliceFlag{
			Name:  "db",
			Usage: "The target `schema(s)` to modify. Comma-separated values are accepted. Supply \"all\" to migrate all schemas.",
			Value: cli.NewStringSlice("all"),
		},
	}

	action := func(cmd *cli.Context) error {
		out := outFactory()

		if cmd.NArg() != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		var schemaNames = cmd.StringSlice("db")

		if len(schemaNames) == 1 || schemaNames[0] == "" {
			schemaNames = nil
		}
		if len(schemaNames) == 1 || schemaNames[0] == "all" {
			schemaNames = schemas.SchemaNames
		}
		if len(schemaNames) == 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a schema via -db"))
			return flag.ErrHelp
		}

		ctx := cmd.Context
		r, err := factory(ctx, schemaNames)
		if err != nil {
			return err
		}

		return r.Validate(ctx, schemaNames...)
	}

	return &cli.Command{
		Name:        "validate",
		Usage:       "Validate the current schema",
		Description: ConstructLongHelp(),
		Flags:       flags,
		Action:      action,
	}
}
