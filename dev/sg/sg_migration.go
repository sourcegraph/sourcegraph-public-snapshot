package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/sourcegraph/run"
	"github.com/urfave/cli/v2"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/migration"
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

	squashInContainer     bool
	squashInContainerFlag = &cli.BoolFlag{
		Name:        "in-container",
		Usage:       "Launch Postgres in a Docker container for squashing; do not use the host",
		Value:       false,
		Destination: &squashInContainer,
	}

	squashInTimescaleDBContainer     bool
	squashInTimescaleDBContainerFlag = &cli.BoolFlag{
		Name:        "in-timescaledb-container",
		Usage:       "Launch TimescaleDB in a Docker container for squashing; do not use the host",
		Value:       false,
		Destination: &squashInTimescaleDBContainer,
	}

	skipTeardown     bool
	skipTeardownFlag = &cli.BoolFlag{
		Name:        "skip-teardown",
		Usage:       "Skip tearing down the database created to run all registered migrations",
		Value:       false,
		Destination: &skipTeardown,
	}

	skipSquashData     bool
	skipSquashDataFlag = &cli.BoolFlag{
		Name:        "skip-data",
		Usage:       "Skip writing data rows into the squashed migration",
		Value:       false,
		Destination: &skipSquashData,
	}

	outputFilepath     string
	outputFilepathFlag = &cli.StringFlag{
		Name:        "f",
		Usage:       "The output filepath",
		Required:    true,
		Destination: &outputFilepath,
	}

	targetRevision     string
	targetRevisionFlag = &cli.StringFlag{
		Name:        "rev",
		Usage:       "The target revision",
		Required:    true,
		Destination: &targetRevision,
	}
)

var (
	addCommand = &cli.Command{
		Name:        "add",
		ArgsUsage:   "<name>",
		Usage:       "Add a new migration file",
		Description: cliutil.ConstructLongHelp(),
		Flags:       []cli.Flag{migrateTargetDatabaseFlag},
		Action:      addExec,
	}

	revertCommand = &cli.Command{
		Name:        "revert",
		ArgsUsage:   "<commit>",
		Usage:       "Revert the migrations defined on the given commit",
		Description: cliutil.ConstructLongHelp(),
		Action:      revertExec,
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
	driftCommand    = cliutil.Drift("sg migration", makeRunner, outputFactory, cliutil.GCSExpectedSchemaFactory, localGitExpectedSchemaFactory)
	addLogCommand   = cliutil.AddLog("sg migration", makeRunner, outputFactory)

	leavesCommand = &cli.Command{
		Name:        "leaves",
		ArgsUsage:   "<commit>",
		Usage:       "Identify the migration leaves for the given commit",
		Description: cliutil.ConstructLongHelp(),
		Action:      leavesExec,
	}

	squashCommand = &cli.Command{
		Name:        "squash",
		ArgsUsage:   "<current-release>",
		Usage:       "Collapse migration files from historic releases together",
		Description: cliutil.ConstructLongHelp(),
		Flags:       []cli.Flag{migrateTargetDatabaseFlag, squashInContainerFlag, squashInTimescaleDBContainerFlag, skipTeardownFlag, skipSquashDataFlag},
		Action:      squashExec,
	}

	squashAllCommand = &cli.Command{
		Name:        "squash-all",
		ArgsUsage:   "",
		Usage:       "Collapse schema definitions into a single SQL file",
		Description: cliutil.ConstructLongHelp(),
		Flags:       []cli.Flag{migrateTargetDatabaseFlag, squashInContainerFlag, squashInTimescaleDBContainerFlag, skipTeardownFlag, skipSquashDataFlag, outputFilepathFlag},
		Action:      squashAllExec,
	}

	visualizeCommand = &cli.Command{
		Name:        "visualize",
		ArgsUsage:   "",
		Usage:       "Output a DOT visualization of the migration graph",
		Description: cliutil.ConstructLongHelp(),
		Flags:       []cli.Flag{migrateTargetDatabaseFlag, outputFilepathFlag},
		Action:      visualizeExec,
	}

	rewriteCommand = &cli.Command{
		Name:        "rewrite",
		ArgsUsage:   "",
		Usage:       "Rewrite schemas definitions as they were at a particular version",
		Description: cliutil.ConstructLongHelp(),
		Flags:       []cli.Flag{migrateTargetDatabaseFlag, targetRevisionFlag},
		Action:      rewriteExec,
	}

	migrationCommand = &cli.Command{
		Name:  "migration",
		Usage: "Modifies and runs database migrations",
		UsageText: `
# Migrate local default database up all the way
sg migration up

# Migrate specific database down one migration
sg migration down --db codeintel

# Add new migration for specific database
sg migration add --db codeintel 'add missing index'

# Squash migrations for default database
sg migration squash
`,
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
			squashAllCommand,
			visualizeCommand,
			rewriteCommand,
		},
	}
)

func makeRunner(ctx context.Context, schemaNames []string) (cliutil.Runner, error) {
	filesystemSchemas, err := getFilesystemSchemas()
	if err != nil {
		return nil, err
	}

	return makeRunnerWithSchemas(ctx, schemaNames, filesystemSchemas)
}

func makeRunnerWithSchemas(ctx context.Context, schemaNames []string, schemas []*schemas.Schema) (cliutil.Runner, error) {
	// Try to read the `sg` configuration so we can read ENV vars from the
	// configuration and use process env as fallback.
	var getEnv func(string) string
	config, _ := getConfig()
	logger := log.Scoped("migrations.runner", "migration runner")
	if config != nil {
		getEnv = config.GetEnv
	} else {
		getEnv = os.Getenv
	}

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext)))
	}
	r, err := connections.RunnerFromDSNsWithSchemas(logger, postgresdsn.RawDSNsBySchema(schemaNames, getEnv), "sg", storeFactory, schemas)
	if err != nil {
		return nil, err
	}

	return cliutil.NewShim(r), nil
}

// localGitExpectedSchemaFactory returns the description of the given schema at the given version via the
// (assumed) local git clone. If the version is not resolvable as a git rev-like, or if the file does not
// exist at that revision, then a false valued-flag is returned. All other failures are reported as errors.
func localGitExpectedSchemaFactory(filename, version string) (schemaDescription schemas.SchemaDescription, _ bool, _ error) {
	ctx := context.Background()
	output := root.Run(run.Cmd(ctx, "git", "show", fmt.Sprintf("%s:%s", version, filename)))

	if err := output.Wait(); err != nil {
		// See if there is an error indicating a missing object, but no other problems
		return schemas.SchemaDescription{}, false, filterLocalGitErrors(filename, version, err)
	}

	if err := json.NewDecoder(output).Decode(&schemaDescription); err != nil {
		return schemas.SchemaDescription{}, false, err
	}

	return schemaDescription, true, nil
}

func filterLocalGitErrors(filename, version string, err error) error {
	if err == nil {
		return nil
	}

	missingMessages := []string{
		// unknown revision
		fmt.Sprintf("fatal: invalid object name '%s'", version),

		// path unknown to the revision (regardless of repo state)
		fmt.Sprintf("fatal: path '%s' does not exist in '%s'", filename, version),
		fmt.Sprintf("fatal: path '%s' exists on disk, but not in '%s'", filename, version),
	}
	for _, missingMessage := range missingMessages {
		if strings.Contains(err.Error(), missingMessage) {
			return nil
		}
	}

	return err
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
	fs, err := db.GetFSForPath(name)()
	if err != nil {
		return nil, err
	}

	schema, err := schemas.ResolveSchema(fs, name)
	if err != nil {
		return nil, errors.Newf("malformed migration definitions: %w", err)
	}

	return schema, nil
}

func addExec(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return cli.NewExitError("no migration name specified", 1)
	}
	if len(args) != 1 {
		return cli.NewExitError("too many arguments", 1)
	}

	var (
		databaseName = migrateTargetDatabase
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		return cli.NewExitError(fmt.Sprintf("database %q not found :(", databaseName), 1)
	}

	return migration.Add(database, args[0])
}

func revertExec(ctx *cli.Context) error {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return cli.NewExitError("no commit specified", 1)
	}
	if len(args) != 1 {
		return cli.NewExitError("too many arguments", 1)
	}

	return migration.Revert(db.Databases(), args[0])
}

func squashExec(ctx *cli.Context) (err error) {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return cli.NewExitError("no current-version specified", 1)
	}
	if len(args) != 1 {
		return cli.NewExitError("too many arguments", 1)
	}

	var (
		databaseName = migrateTargetDatabase
		database, ok = db.DatabaseByName(databaseName)
	)
	if !ok {
		return cli.NewExitError(fmt.Sprintf("database %q not found :(", databaseName), 1)
	}

	// Get the last migration that existed in the version _before_ `minimumMigrationSquashDistance` releases ago
	commit, err := findTargetSquashCommit(args[0])
	if err != nil {
		return err
	}
	std.Out.Writef("Squashing migration files defined up through %s", commit)

	return migration.Squash(database, commit, squashInContainer || squashInTimescaleDBContainer, squashInTimescaleDBContainer, skipTeardown, skipSquashData)
}

func visualizeExec(ctx *cli.Context) (err error) {
	args := ctx.Args().Slice()
	if len(args) != 0 {
		return cli.NewExitError("too many arguments", 1)
	}

	if outputFilepath == "" {
		return cli.NewExitError("Supply an output file with -f", 1)
	}

	var (
		databaseName = migrateTargetDatabase
		database, ok = db.DatabaseByName(databaseName)
	)

	if !ok {
		return cli.NewExitError(fmt.Sprintf("database %q not found :(", databaseName), 1)
	}

	return migration.Visualize(database, outputFilepath)
}

func rewriteExec(ctx *cli.Context) (err error) {
	args := ctx.Args().Slice()
	if len(args) != 0 {
		return cli.NewExitError("too many arguments", 1)
	}

	if targetRevision == "" {
		return cli.NewExitError("Supply a target revision with -rev", 1)
	}

	var (
		databaseName = migrateTargetDatabase
		database, ok = db.DatabaseByName(databaseName)
	)

	if !ok {
		return cli.NewExitError(fmt.Sprintf("database %q not found :(", databaseName), 1)
	}

	return migration.Rewrite(database, targetRevision)
}

func squashAllExec(ctx *cli.Context) (err error) {
	args := ctx.Args().Slice()
	if len(args) != 0 {
		return cli.NewExitError("too many arguments", 1)
	}

	if outputFilepath == "" {
		return cli.NewExitError("Supply an output file with -f", 1)
	}

	var (
		databaseName = migrateTargetDatabase
		database, ok = db.DatabaseByName(databaseName)
	)

	if !ok {
		return cli.NewExitError(fmt.Sprintf("database %q not found :(", databaseName), 1)
	}

	return migration.SquashAll(database, squashInContainer || squashInTimescaleDBContainer, squashInTimescaleDBContainer, skipTeardown, skipSquashData, outputFilepath)
}

func leavesExec(ctx *cli.Context) (err error) {
	args := ctx.Args().Slice()
	if len(args) == 0 {
		return cli.NewExitError("no commit specified", 1)
	}
	if len(args) != 1 {
		return cli.NewExitError("too many arguments", 1)
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
