package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/migration"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/cliutil"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/schemas"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/database/postgresdsn"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	addFlagSet          = flag.NewFlagSet("sg migration add", flag.ExitOnError)
	addDatabaseNameFlag = addFlagSet.String("db", db.DefaultDatabase.Name, "The target schema to modify")
	addCommand          = &ffcli.Command{
		Name:       "add",
		ShortUsage: fmt.Sprintf("sg migration add [-db=%s] <name>", db.DefaultDatabase.Name),
		ShortHelp:  "Add a new migration file",
		FlagSet:    addFlagSet,
		Exec:       addExec,
		LongHelp:   cliutil.ConstructLongHelp(),
	}

	migrationRevertFlagSet = flag.NewFlagSet("sg migration revert", flag.ExitOnError)
	revertCommand          = &ffcli.Command{
		Name:       "revert",
		ShortUsage: "sg migration revert <commit>",
		ShortHelp:  "Revert the migrations defined on the given commit",
		FlagSet:    migrationRevertFlagSet,
		Exec:       revertExec,
		LongHelp:   cliutil.ConstructLongHelp(),
	}

	upCommand       = cliutil.Up("sg migration", makeRunner, stdout.Out, true)
	upToCommand     = cliutil.UpTo("sg migration", makeRunner, stdout.Out, true)
	UndoCommand     = cliutil.Undo("sg migration", makeRunner, stdout.Out, true)
	downToCommand   = cliutil.DownTo("sg migration", makeRunner, stdout.Out, true)
	validateCommand = cliutil.Validate("sg validate", makeRunner, stdout.Out)
	addLogCommand   = cliutil.AddLog("sg migration", makeRunner, stdout.Out)

	leavesFlagSet = flag.NewFlagSet("sg migration leaves", flag.ExitOnError)
	leavesCommand = &ffcli.Command{
		Name:       "leaves",
		ShortUsage: "sg migration leaves <commit>",
		ShortHelp:  "Identiy the migration leaves for the given commit",
		FlagSet:    leavesFlagSet,
		Exec:       leavesExec,
		LongHelp:   cliutil.ConstructLongHelp(),
	}

	squashFlagSet          = flag.NewFlagSet("sg migration squash", flag.ExitOnError)
	squashDatabaseNameFlag = squashFlagSet.String("db", db.DefaultDatabase.Name, "The target schema to modify")
	squashCommand          = &ffcli.Command{
		Name:       "squash",
		ShortUsage: fmt.Sprintf("sg migration squash [-db=%s] <current-release>", db.DefaultDatabase.Name),
		ShortHelp:  "Collapse migration files from historic releases together",
		FlagSet:    squashFlagSet,
		Exec:       squashExec,
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
			addCommand,
			revertCommand,
			upCommand,
			upToCommand,
			UndoCommand,
			downToCommand,
			validateCommand,
			addLogCommand,
			leavesCommand,
			squashCommand,
		},
	}
)

func makeRunner(ctx context.Context, schemaNames []string) (cliutil.Runner, error) {
	// Try to read the `sg` configuration so we can read ENV vars from the
	// configuration and use process env as fallback.
	var getEnv func(string) string
	ok, _ := parseConf(*configFlag, *overwriteConfigFlag)
	if ok {
		getEnv = globalConf.GetEnv
	} else {
		getEnv = os.Getenv
	}

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext)))
	}
	schemas, err := getFilesystemSchemas()
	if err != nil {
		return nil, err
	}
	r, err := connections.RunnerFromDSNsWithSchemas(postgresdsn.RawDSNsBySchema(schemaNames, getEnv), "sg", storeFactory, schemas)
	if err != nil {
		return nil, err
	}

	return cliutil.NewShim(r), nil
}

func getFilesystemSchemas() (schemas []*schemas.Schema, errs error) {
	for _, name := range []string{"frontend", "codeintel", "codeinsights"} {
		schema, err := resolveSchema(name)
		if err != nil {
			errs = errors.Append(errs, errors.Newf("%s: %w", name, err))
		} else {
			schemas = append(schemas, schema)
		}
	}
	return
}

func resolveSchema(name string) (*schemas.Schema, error) {
	repositoryRoot, err := root.RepositoryRoot()
	if err != nil {
		if errors.Is(err, root.ErrNotInsideSourcegraph) {
			return nil, errors.Newf("sg migration command uses the migrations defined on the local filesystem: %w", err)
		}
		return nil, err
	}

	schema, err := schemas.ResolveSchema(os.DirFS(filepath.Join(repositoryRoot, "migrations", name)), name)
	if err != nil {
		return nil, errors.Newf("malformed migration definitions: %w", err)
	}

	return schema, nil
}

func addExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No migration name specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName = *addDatabaseNameFlag
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	return migration.Add(database, args[0])
}

func revertExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No commit specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	return migration.Revert(db.Databases(), args[0])
}

func squashExec(ctx context.Context, args []string) (err error) {
	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No current-version specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName = *squashDatabaseNameFlag
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	// Get the last migration that existed in the version _before_ `minimumMigrationSquashDistance` releases ago
	commit, err := findTargetSquashCommit(args[0])
	if err != nil {
		return err
	}
	stdout.Out.Writef("Squashing migration files defined up through %s", commit)

	return migration.Squash(database, commit)
}

func leavesExec(ctx context.Context, args []string) (err error) {
	if len(args) == 0 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "No commit specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		stdout.Out.WriteLine(output.Linef("", output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	return migration.LeavesForCommit(db.Databases(), args[0])
}

// minimumMigrationSquashDistance is the minimum number of releases a migration is guaranteed to exist
// as a non-squashed file.
//
// A squash distance of 1 will allow one minor downgrade.
// A squash distance of 2 will allow two minor downgrades.
// etc
const minimumMigrationSquashDistance = 2

// findTargetSquashCommit constructs the git version tag that is `minimumMIgrationSquashDistance` minor
// releases ago.
func findTargetSquashCommit(migrationName string) (string, error) {
	currentVersion, err := semver.NewVersion(migrationName)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("v%d.%d.0", currentVersion.Major(), currentVersion.Minor()-minimumMigrationSquashDistance-1), nil
}
