package cliutil

import (
	"context"
	"flag"
	"fmt"
	"strconv"

	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

func AddLog(commandName string, factory RunnerFactory, out *output.Output) *ffcli.Command {
	var (
		flagSet        = flag.NewFlagSet(fmt.Sprintf("%s add-log", commandName), flag.ExitOnError)
		schemaNameFlag = flagSet.String("db", "", `The target schema to modify.`)
		versionFlag    = flagSet.String("version", "", "The migration version.")
		upFlag         = flagSet.Bool("up", true, "The migration direction.")
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

		if *versionFlag == "" {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: supply a migration version via -version"))
			return flag.ErrHelp
		}
		version, err := strconv.Atoi(*versionFlag)
		if err != nil {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: invalid migration version %q", *versionFlag))
			return flag.ErrHelp
		}

		r, err := factory(ctx, []string{*schemaNameFlag})
		if err != nil {
			return err
		}
		store, err := r.Store(ctx, *schemaNameFlag)
		if err != nil {
			return err
		}

		return store.WithMigrationLog(ctx, definition.Definition{ID: version}, *upFlag, noop)
	}

	return &ffcli.Command{
		Name:       "add-log",
		ShortUsage: fmt.Sprintf("%s add-log -db=<schema> -version=<version> [-up=true|false]", commandName),
		ShortHelp:  "Add an entry to the migration log",
		FlagSet:    flagSet,
		Exec:       exec,
		LongHelp:   ConstructLongHelp(),
	}
}

func noop() error {
	return nil
}
