package cliutil

import (
	"flag"
	"fmt"

	"github.com/inconshreveable/log15"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func AddLog(commandName string, factory RunnerFactory, outFactory OutputFactory) *cli.Command {
	flags := []cli.Flag{
		&cli.StringFlag{
			Name:     "db",
			Usage:    "The target `schema` to modify.",
			Required: true,
		},
		&cli.IntFlag{
			Name:     "version",
			Usage:    "The migration `version` to log.",
			Required: true,
		},
		&cli.BoolFlag{
			Name:  "up",
			Usage: "The migration direction.",
			Value: true,
		},
	}

	action := func(cmd *cli.Context) error {
		out := outFactory()

		if cmd.NArg() != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		var (
			schemaNameFlag = cmd.String("db")
			versionFlag    = cmd.Int("version")
			upFlag         = cmd.Bool("up")
		)

		ctx := cmd.Context
		r, err := factory(ctx, []string{schemaNameFlag})
		if err != nil {
			return err
		}
		store, err := r.Store(ctx, schemaNameFlag)
		if err != nil {
			return err
		}

		log15.Info("Writing new completed migration log", "schema", schemaNameFlag, "version", versionFlag, "up", upFlag)
		return store.WithMigrationLog(ctx, definition.Definition{ID: versionFlag}, upFlag, noop)
	}

	return &cli.Command{
		Name:        "add-log",
		UsageText:   fmt.Sprintf("%s add-log -db=<schema> -version=<version> [-up=true|false]", commandName),
		Usage:       "Add an entry to the migration log",
		Description: ConstructLongHelp(),
		Flags:       flags,
		Action:      action,
	}
}

func noop() error {
	return nil
}
