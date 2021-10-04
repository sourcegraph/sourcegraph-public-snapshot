package main

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/migration"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/squash"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	migrationAddFlagSet          = flag.NewFlagSet("sg migration add", flag.ExitOnError)
	migrationAddDatabaseNameFlag = migrationAddFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	migrationAddCommand          = &ffcli.Command{
		Name:       "add",
		ShortUsage: fmt.Sprintf("sg migration add [-db=%s] <name>", db.DefaultDatabase.Name),
		ShortHelp:  "Add a new migration file",
		FlagSet:    migrationAddFlagSet,
		Exec:       migrationAddExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationUpFlagSet          = flag.NewFlagSet("sg migration up", flag.ExitOnError)
	migrationUpDatabaseNameFlag = migrationUpFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	migrationUpNFlag            = migrationUpFlagSet.Int("n", 1, "How many migrations to apply.")
	migrationUpCommand          = &ffcli.Command{
		Name:       "up",
		ShortUsage: fmt.Sprintf("sg migration up [-db=%s] [-n]", db.DefaultDatabase.Name),
		ShortHelp:  "Run up migration files",
		FlagSet:    migrationUpFlagSet,
		Exec:       migrationUpExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationDownFlagSet          = flag.NewFlagSet("sg migration down", flag.ExitOnError)
	migrationDownDatabaseNameFlag = migrationDownFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance.")
	migrationDownNFlag            = migrationDownFlagSet.Int("n", 1, "How many migrations to apply.")
	migrationDownCommand          = &ffcli.Command{
		Name:       "down",
		ShortUsage: fmt.Sprintf("sg migration down [-db=%s] [-n=1]", db.DefaultDatabase.Name),
		ShortHelp:  "Run down migration files",
		FlagSet:    migrationDownFlagSet,
		Exec:       migrationDownExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationSquashFlagSet          = flag.NewFlagSet("sg migration squash", flag.ExitOnError)
	migrationSquashDatabaseNameFlag = migrationSquashFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance")
	migrationSquashCommand          = &ffcli.Command{
		Name:       "squash",
		ShortUsage: fmt.Sprintf("sg migration squash [-db=%s] <current-release>", db.DefaultDatabase.Name),
		ShortHelp:  "Collapse migration files from historic releases together",
		FlagSet:    migrationSquashFlagSet,
		Exec:       migrationSquashExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationFixupFlagSet          = flag.NewFlagSet("sg migration fixup", flag.ExitOnError)
	migrationFixupDatabaseNameFlag = migrationFixupFlagSet.String("db", "all", "The target database instance (or 'all' for all databases)")
	migrationFixupMainNameFlag     = migrationFixupFlagSet.String("main", "main", "The branch/revision to compare with")
	migrationFixupRunFlag          = migrationFixupFlagSet.Bool("run", true, "Run the migrations in your local database")
	migrationFixupCommand          = &ffcli.Command{
		Name:       "fixup",
		ShortUsage: fmt.Sprintf("sg migration fixup [-db=%s] [-main=%s] [-run=true]", "all", "main"),
		ShortHelp:  "Find and fix any conflicting migration names from rebasing on main. Also properly migrates your local database",
		FlagSet:    migrationFixupFlagSet,
		Exec:       migrationFixupExec,
		LongHelp:   constructMigrationSubcmdLongHelp(),
	}

	migrationFlagSet = flag.NewFlagSet("sg migration", flag.ExitOnError)
	migrationCommand = &ffcli.Command{
		Name:       "migration",
		ShortUsage: "sg migration <command>",
		ShortHelp:  "Modifies and runs database migrations",
		FlagSet:    migrationFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			return flag.ErrHelp
		},
		Subcommands: []*ffcli.Command{
			migrationAddCommand,
			migrationUpCommand,
			migrationDownCommand,
			migrationSquashCommand,
			migrationFixupCommand,
		},
	}
)

func constructMigrationSubcmdLongHelp() string {
	var out strings.Builder

	fmt.Fprintf(&out, "AVAILABLE DATABASES\n")
	var names []string
	for _, name := range db.DatabaseNames() {
		names = append(names, fmt.Sprintf("  %s", name))
	}
	fmt.Fprint(&out, strings.Join(names, "\n"))

	return out.String()
}

func migrationAddExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No migration name specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName  = *migrationAddDatabaseNameFlag
		migrationName = args[0]
		database, ok  = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	upFile, downFile, err := migration.RunAdd(database, migrationName)
	if err != nil {
		return err
	}

	block := out.Block(output.Linef("", output.StyleBold, "Migration files created"))
	block.Writef("Up migration: %s", upFile)
	block.Writef("Down migration: %s", downFile)
	block.Close()

	return nil
}

func migrationUpExec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName = *migrationUpDatabaseNameFlag
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	var n *int
	migrationUpFlagSet.Visit(func(f *flag.Flag) {
		if f.Name == "n" {
			n = migrationUpNFlag
		}
	})

	// Only pass the value of n here if the user actually set it
	// We have to do the dance above because the flags package
	// requires you to define a default value for each flag.
	return migration.RunUp(database, n)
}

func migrationDownExec(ctx context.Context, args []string) error {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName = *migrationDownDatabaseNameFlag
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	return migration.RunDown(database, migrationDownNFlag)
}

// minimumMigrationSquashDistance is the minimum number of releases a migration is guaranteed to exist
// as a non-squashed file.
//
// A squash distance of 1 will allow one minor downgrade.
// A squash distance of 2 will allow two minor downgrades.
// etc
const minimumMigrationSquashDistance = 2

func migrationSquashExec(ctx context.Context, args []string) (err error) {
	if len(args) == 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "No current-version specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName  = *migrationSquashDatabaseNameFlag
		migrationName = args[0]
		database, ok  = db.DatabaseByName(databaseName)
	)
	if !ok {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	currentVersion, err := semver.NewVersion(migrationName)
	if err != nil {
		return err
	}

	// Get the last migration that existed in the version _before_ `minimumMigrationSquashDistance` releases ago
	commit := fmt.Sprintf("v%d.%d.0", currentVersion.Major(), currentVersion.Minor()-minimumMigrationSquashDistance-1)
	out.Writef("Squashing migration files defined up through %s", commit)

	return squash.Run(database, commit)
}

func migrationFixupExec(ctx context.Context, args []string) (err error) {
	if len(args) != 0 {
		out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	branchName := *migrationFixupMainNameFlag
	if branchName == "" {
		branchName = "main"
	}

	databaseName := *migrationFixupDatabaseNameFlag
	if databaseName == "all" {
		for _, databaseName := range db.DatabaseNames() {
			database, _ := db.DatabaseByName(databaseName)
			if err := migration.RunFixup(database, branchName, *migrationFixupRunFlag); err != nil {
				return err
			}
		}

		return nil
	} else {
		database, ok := db.DatabaseByName(databaseName)
		if !ok {
			out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
			return flag.ErrHelp
		}

		return migration.RunFixup(database, branchName, *migrationFixupRunFlag)
	}
}
