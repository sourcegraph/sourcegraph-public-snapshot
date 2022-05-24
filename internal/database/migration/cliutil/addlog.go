package cliutil

import (
	"context"
	"fmt"

	"github.com/inconshreveable/log15"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func AddLog(commandName string, factory RunnerFactory, outFactory OutputFactory) *cli.Command {
	schemaNameFlag := &cli.StringFlag{
		Name:     "db",
		Usage:    "The target `schema` to modify.",
		Required: true,
	}
	versionFlag := &cli.IntFlag{
		Name:     "version",
		Usage:    "The migration `version` to log.",
		Required: true,
	}
	upFlag := &cli.BoolFlag{
		Name:  "up",
		Usage: "The migration direction.",
		Value: true,
	}

	action := makeAction(outFactory, func(ctx context.Context, cmd *cli.Context, out *output.Output) error {
		var (
			schemaName  = schemaNameFlag.Get(cmd)
			versionFlag = versionFlag.Get(cmd)
			upFlag      = upFlag.Get(cmd)
		)

		_, store, err := setupStore(ctx, factory, schemaName)
		if err != nil {
			return err
		}

		log15.Info("Writing new completed migration log", "schema", schemaName, "version", versionFlag, "up", upFlag)
		return store.WithMigrationLog(ctx, definition.Definition{ID: versionFlag}, upFlag, func() error { return nil })
	})

	return &cli.Command{
		Name:        "add-log",
		UsageText:   fmt.Sprintf("%s add-log -db=<schema> -version=<version> [-up=true|false]", commandName),
		Usage:       "Add an entry to the migration log",
		Description: ConstructLongHelp(),
		Action:      action,
		Flags: []cli.Flag{
			schemaNameFlag,
			versionFlag,
			upFlag,
		},
	}
}
