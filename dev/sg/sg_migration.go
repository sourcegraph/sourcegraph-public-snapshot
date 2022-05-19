package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Masterminds/semver"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/migration"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/sgconf"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/std"
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
	migrateTargetDatabase     string
	migrateTargetDatabaseFlag = &cli.StringFlag{
		Name:        "db",
		Usage:       "The target database `schema` to modify",
		Value:       db.DefaultDatabase.Name,
		Destination: &migrateTargetDatabase,
	}
)

var (
	addCommand = &cli.Command{
		Name:        "add",
		ArgsUsage:   "<name>",
		Usage:       "Add a new migration file",
		Description: cliutil.ConstructLongHelp(),
		Flags:       []cli.Flag{migrateTargetDatabaseFlag},
		Action:      execAdapter(addExec),
	}

	revertCommand = &cli.Command{
		Name:        "revert",
		ArgsUsage:   "<commit>",
		Usage:       "Revert the migrations defined on the given commit",
		Description: cliutil.ConstructLongHelp(),
		Action:      execAdapter(revertExec),
	}

	// outputFactory lazily retrieves the global output that might not yet be instantiated
	// at compile-time in sg.
	outputFactory = func() *output.Output { return std.Out.Output }

	upCommand       = cliutil.Up("sg migration", makeRunner, outputFactory, true)
	upToCommand     = cliutil.UpTo("sg migration", makeRunner, outputFactory, true)
	undoCommand     = cliutil.Undo("sg migration", makeRunner, outputFactory, true)
	downToCommand   = cliutil.DownTo("sg migration", makeRunner, outputFactory, true)
	validateCommand = cliutil.Validate("sg migration", makeRunner, outputFactory)
	describeCommand = cliutil.Describe("sg migration", makeRunner, outputFactory)
	driftCommand    = cliutil.Drift("sg migration", makeRunner, outputFactory)
	addLogCommand   = cliutil.AddLog("sg migration", makeRunner, outputFactory)

	leavesCommand = &cli.Command{
		Name:        "leaves",
		ArgsUsage:   "<commit>",
		Usage:       "Identiy the migration leaves for the given commit",
		Description: cliutil.ConstructLongHelp(),
		Action:      execAdapter(leavesExec),
	}

	squashCommand = &cli.Command{
		Name:        "squash",
		ArgsUsage:   "<current-release>",
		Usage:       "Collapse migration files from historic releases together",
		Description: cliutil.ConstructLongHelp(),
		Flags:       []cli.Flag{migrateTargetDatabaseFlag},
		Action:      execAdapter(squashExec),
	}

	migrationCommand = &cli.Command{
		Name:     "migration",
		Usage:    "Modifies and runs database migrations",
		Category: CategoryDev,
		Subcommands: []*cli.Command{
			addCommand,
			revertCommand,
			upCommand,
			upToCommand,
			undoCommand,
			downToCommand,
			validateCommand,
			describeCommand,
			driftCommand,
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
	config, _ := sgconf.Get(configFile, configOverwriteFile)
	if config != nil {
		getEnv = config.GetEnv
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
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No migration name specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName = migrateTargetDatabase
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	return migration.Add(database, args[0])
}

func revertExec(ctx context.Context, args []string) error {
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No commit specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	return migration.Revert(db.Databases(), args[0])
}

func squashExec(ctx context.Context, args []string) (err error) {
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No current-version specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: too many arguments"))
		return flag.ErrHelp
	}

	var (
		databaseName = migrateTargetDatabase
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		std.Out.WriteLine(output.Styledf(output.StyleWarning, "ERROR: database %q not found :(", databaseName))
		return flag.ErrHelp
	}

	// Get the last migration that existed in the version _before_ `minimumMigrationSquashDistance` releases ago
	commit, err := findTargetSquashCommit(args[0])
	if err != nil {
		return err
	}
	std.Out.Writef("Squashing migration files defined up through %s", commit)

	return migration.Squash(database, commit)
}

func leavesExec(ctx context.Context, args []string) (err error) {
	if len(args) == 0 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "No commit specified"))
		return flag.ErrHelp
	}
	if len(args) != 1 {
		std.Out.WriteLine(output.Styled(output.StyleWarning, "ERROR: too many arguments"))
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
