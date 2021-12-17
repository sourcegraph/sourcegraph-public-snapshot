package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"

	"github.com/Masterminds/semver"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/migration"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/squash"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
		LongHelp:   cliutil.ConstructLongHelp(),
	}

	upCommand   = cliutil.Up("sg migration", runMigration, stdout.Out)
	downCommand = cliutil.Down("sg migration", runMigration, stdout.Out)

	migrationSquashFlagSet          = flag.NewFlagSet("sg migration squash", flag.ExitOnError)
	migrationSquashDatabaseNameFlag = migrationSquashFlagSet.String("db", db.DefaultDatabase.Name, "The target database instance")
	migrationSquashCommand          = &ffcli.Command{
		Name:       "squash",
		ShortUsage: fmt.Sprintf("sg migration squash [-db=%s] <current-release>", db.DefaultDatabase.Name),
		ShortHelp:  "Collapse migration files from historic releases together",
		FlagSet:    migrationSquashFlagSet,
		Exec:       migrationSquashExec,
		LongHelp:   cliutil.ConstructLongHelp(),
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
			upCommand,
			downCommand,
			migrationSquashCommand,
		},
	}
)

func runMigration(ctx context.Context, options runner.Options) error {
	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext))
	}

	return connections.RunnerFromDSNs(postgresdsn.RawDSNsBySchema(), "sg", storeFactory).Run(ctx, options)
}

func migrationAddExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No migration name specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName  = *migrationAddDatabaseNameFlag
		migrationName = args[0]
		database, ok  = db.DatabaseByName(databaseName)
	)
	if !ok {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	upFile, downFile, err := migration.RunAdd(database, migrationName)
	if err != nil {
		return err
	}

	block := stdout.Out.Block(output.Linef("", output.StyleBold, "Migration files created"))
	block.Writef("Up migration: %s", upFile)
	block.Writef("Down migration: %s", downFile)
	block.Close()

	return nil
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
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No current-version specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName  = *migrationSquashDatabaseNameFlag
		migrationName = args[0]
		database, ok  = db.DatabaseByName(databaseName)
	)
	if !ok {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	currentVersion, err := semver.NewVersion(migrationName)
	if err != nil {
		return err
	}

	// Get the last migration that existed in the version _before_ `minimumMigrationSquashDistance` releases ago
	commit := fmt.Sprintf("v%d.%d.0", currentVersion.Major(), currentVersion.Minor()-minimumMigrationSquashDistance-1)
	stdout.Out.Writef("Squashing migration files defined up through %s", commit)

	return squash.Run(database, commit)
}
