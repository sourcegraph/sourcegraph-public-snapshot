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

type RunFunc func(ctx context.Context, options runner.Options) error

func Flags(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	rootFlagSet := flag.NewFlagSet(commandName, flag.ExitOnError)

	return &ffcli.Command{
		Name:       commandName,
		ShortUsage: fmt.Sprintf("%s <command>", commandName),
		ShortHelp:  "Modifies and runs database migrations",
		FlagSet:    rootFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			Up(commandName, run, out),
			Down(commandName, run, out),
		},
	}
}

func Up(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	var (
		upFlagSet          = flag.NewFlagSet(fmt.Sprintf("%s up", commandName), flag.ExitOnError)
		upDatabaseNameFlag = upFlagSet.String("db", "all", `The target database instance. Supply "all" (the default) to migrate all databases.`)
		upNFlag            = upFlagSet.Int("n", 0, "How many migrations to apply. Zero (the default) applies all migrations.")
	)

	execUp := func(ctx context.Context, args []string) error {
		if len(args) != 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
			return flag.ErrHelp
		}

		if *upDatabaseNameFlag == "all" && *upNFlag != 0 {
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
			Up:            true,
			NumMigrations: *upNFlag,
			SchemaNames:   databaseNames,
		})
	}

	return &ffcli.Command{
		Name:       "up",
		ShortUsage: fmt.Sprintf("%s up [-db=all] [-n=0]", commandName),
		ShortHelp:  "Run up migrations",
		FlagSet:    upFlagSet,
		Exec:       execUp,
		LongHelp:   ConstructLongHelp(),
	}
}

func Down(commandName string, run RunFunc, out *output.Output) *ffcli.Command {
	var (
		downFlagSet          = flag.NewFlagSet(fmt.Sprintf("%s down", commandName), flag.ExitOnError)
		downDatabaseNameFlag = downFlagSet.String("db", "", "The target database instance.")
		downNFlag            = downFlagSet.Int("n", 1, "How many migrations to apply.")
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

		if *downNFlag == 0 {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: invalid number of migrations"))
			return flag.ErrHelp
		}

		return run(ctx, runner.Options{
			Up:            false,
			NumMigrations: *downNFlag,
			SchemaNames:   []string{*downDatabaseNameFlag},
		})
	}

	return &ffcli.Command{
		Name:       "down",
		ShortUsage: fmt.Sprintf("%s down -db=... [-n=1]", commandName),
		ShortHelp:  "Run down migrations",
		FlagSet:    downFlagSet,
		Exec:       execDown,
		LongHelp:   ConstructLongHelp(),
	}
}

func ConstructLongHelp() string {
	names := make([]string, 0, len(schemas.SchemaNames))
	for _, name := range schemas.SchemaNames {
		names = append(names, fmt.Sprintf("  %s", name))
	}

	return fmt.Sprintf("AVAILABLE SCHEMAS\n%s", strings.Join(names, "\n"))
}
