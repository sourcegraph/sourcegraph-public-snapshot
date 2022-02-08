package migration

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
	"github.com/sourcegraph/sourcegraph/dev/sg/internal/stdout"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/runner"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

const (
	squasherContainerName         = "squasher"
	squasherContainerExposedPort  = 5433
	squasherContainerPostgresName = "postgres"
)

func Squash(database db.Database, commit string) error {
	definitions, err := readDefinitions(database)
	if err != nil {
		return err
	}

	newRoot, ok, err := selectNewRootMigration(database, definitions, commit)
	if err != nil {
		return err
	}
	if !ok {
		return errors.Newf("no migrations exist at commit %s", commit)
	}

	// Run migrations up to the new selected root and dump the database into a single migration file pair
	squashedUpMigration, squashedDownMigration, err := generateSquashedMigrations(database, []int{newRoot})
	if err != nil {
		return err
	}

	stdout.Out.Write("")
	block := stdout.Out.Block(output.Linef("", output.StyleBold, "Updated filesystem"))
	defer block.Close()

	upPath, downPath, metadataPath, err := makeMigrationFilenames(database, newRoot)
	if err != nil {
		return err
	}

	contents := map[string]string{
		upPath:       squashedUpMigration,
		downPath:     squashedDownMigration,
		metadataPath: "name: 'squashed migrations'\n",
	}

	if err := writeMigrationFiles(contents); err != nil {
		return err
	}

	block.Writef("Created: %s", upPath)
	block.Writef("Created: %s", downPath)
	block.Writef("Created: %s", metadataPath)

	// Remove the migration file pairs that were just squashed
	filenames, err := removeAncestorsOf(database, definitions, newRoot)
	if err != nil {
		return err
	}

	for _, filename := range filenames {
		block.Writef("Deleted: %s", filename)
	}

	return nil
}

// selectNewRootMigration selects the most recently defined migration that dominates the leaf
// migrations of the schema at the given commit. This ensures that whenever we squash migrations,
// we do so between a portion of the graph with a single entry and a single exit, which can
// be easily collapsible into one file that can replace an existing migration node in-place.
func selectNewRootMigration(database db.Database, ds *definition.Definitions, commit string) (int, bool, error) {
	migrationsDir := filepath.Join("migrations", database.Name)

	output, err := run.GitCmd("ls-tree", "-r", "--name-only", commit, migrationsDir)
	if err != nil {
		return 0, false, err
	}
	lines := strings.Split(output, "\n")

	versions := make([]int, 0, len(lines))
	for _, filename := range lines {
		if version, err := strconv.Atoi(strings.Split(filename, "_")[0]); err == nil {
			versions = append(versions, version)
		}
	}

	ds, err = ds.Filter(versions)
	if err != nil {
		return 0, false, err
	}

	id, ok := ds.LeafDominator()
	if !ok {
		return 0, false, nil
	}

	return id.ID, true, nil
}

// generateSquashedMigrations generates the content of a migration file pair that contains the contents
// of a database up to a given migration index. This function will launch a daemon Postgres container,
// migrate a fresh database up to the given migration index, then dump and sanitize the contents.
func generateSquashedMigrations(database db.Database, targetVersions []int) (up, down string, err error) {
	postgresDSN := fmt.Sprintf(
		"postgres://postgres@127.0.0.1:%d/%s?sslmode=disable",
		squasherContainerExposedPort,
		database.Name,
	)

	teardown, err := runPostgresContainer(database.Name)
	if err != nil {
		return "", "", err
	}
	defer func() {
		err = teardown(err)
	}()

	if err := runTargetedUpMigrations(database, targetVersions, postgresDSN); err != nil {
		return "", "", err
	}

	upMigration, err := generateSquashedUpMigration(database, postgresDSN)
	if err != nil {
		return "", "", err
	}

	return upMigration, "-- Nothing\n", nil
}

// runTargetedUpMigrations runs up migration targeting the given versions on the given database instance.
func runTargetedUpMigrations(database db.Database, targetVersions []int, postgresDSN string) (err error) {
	pending := stdout.Out.Pending(output.Line("", output.StylePending, "Migrating PostgreSQL schema..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Migrated PostgreSQL schema"))
		} else {
			pending.Destroy()
		}
	}()

	ctx := context.Background()

	storeFactory := func(db *sql.DB, migrationsTable string) connections.Store {
		return connections.NewStoreShim(store.NewWithDB(db, migrationsTable, store.NewOperations(&observation.TestContext)))
	}

	options := runner.Options{
		Operations: []runner.MigrationOperation{
			{
				SchemaName:     database.Name,
				Type:           runner.MigrationOperationTypeTargetedUp,
				TargetVersions: targetVersions,
			},
		},
	}

	dsns := map[string]string{
		database.Name: postgresDSN,
	}

	return connections.RunnerFromDSNs(dsns, "sg", storeFactory).Run(ctx, options)
}

// runPostgresContainer runs a postgres:12.6 daemon with an empty db with the given name.
// This method returns a teardown function that filters the error value of the calling
// function, as well as any immediate synchronous error.
func runPostgresContainer(databaseName string) (_ func(err error) error, err error) {
	pending := stdout.Out.Pending(output.Line("", output.StylePending, "Starting PostgreSQL 12 in a container..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Started PostgreSQL in a container"))
		} else {
			pending.Destroy()
		}
	}()

	teardown := func(err error) error {
		killArgs := []string{
			"kill",
			squasherContainerName,
		}
		if _, killErr := run.DockerCmd(killArgs...); killErr != nil {
			err = errors.Append(err, errors.Newf("failed to stop docker container: %s", killErr))
		}

		return err
	}

	runArgs := []string{
		"run",
		"--rm", "-d",
		"--name", squasherContainerName,
		"-p", fmt.Sprintf("%d:5432", squasherContainerExposedPort),
		"-e", "POSTGRES_HOST_AUTH_METHOD=trust",
		"postgres:12.6",
	}
	if _, err := run.DockerCmd(runArgs...); err != nil {
		return nil, err
	}

	// TODO - check health instead
	pending.Write("Waiting for container to start up...")
	time.Sleep(5 * time.Second)
	pending.Write("PostgreSQL is accepting connections")

	execArgs := []string{
		"exec",
		"-u", "postgres",
		squasherContainerName,
		"createdb", databaseName,
	}
	if _, err := run.DockerCmd(execArgs...); err != nil {
		return nil, teardown(err)
	}

	return teardown, nil
}

// generateSquashedUpMigration returns the contents of an up migration file containing the
// current contents of the given database.
func generateSquashedUpMigration(database db.Database, postgresDSN string) (_ string, err error) {
	pending := stdout.Out.Pending(output.Line("", output.StylePending, "Dumping current database..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Dumped current database"))
		} else {
			pending.Destroy()
		}
	}()

	pgDump := func(args ...string) (string, error) {
		cmd := exec.Command("pg_dump", append([]string{postgresDSN}, args...)...)
		return run.InRoot(cmd)
	}

	excludeTables := []string{
		"*schema_migrations",
		"migration_logs",
		"migration_logs_id_seq",
	}

	args := []string{
		"--schema-only",
		"--no-owner",
	}
	for _, tableName := range excludeTables {
		args = append(args, "--exclude-table", tableName)
	}

	pgDumpOutput, err := pgDump(args...)
	if err != nil {
		return "", err
	}

	for _, table := range database.DataTables {
		dataOutput, err := pgDump("--data-only", "--inserts", "--table", table)
		if err != nil {
			return "", err
		}

		pgDumpOutput += dataOutput
	}

	return sanitizePgDumpOutput(pgDumpOutput), nil
}

var (
	migrationDumpRemovePrefixes = []string{
		"--",                                    // remove comments
		"SET ",                                  // remove settings header
		"SELECT pg_catalog.set_config",          // remove settings header
		`could not find a "pg_dump" to execute`, // remove common warning from docker container
		"DROP EXTENSION ",                       // do not drop extensions if they already exist
	}

	migrationDumpRemovePatterns = map[*regexp.Regexp]string{
		regexp.MustCompile(`\bpublic\.`):              "",
		regexp.MustCompile(`\s*WITH SCHEMA public\b`): "",
		regexp.MustCompile(`\n{3,}`):                  "\n\n",
	}
)

// sanitizePgDumpOutput sanitizes the output of pg_dump and wraps the content in a
// transaction block to fit the style of our other migrations.
func sanitizePgDumpOutput(content string) string {
	lines := strings.Split(content, "\n")

	filtered := lines[:0]
outer:
	for _, line := range lines {
		for _, prefix := range migrationDumpRemovePrefixes {
			if strings.HasPrefix(line, prefix) {
				continue outer
			}
		}

		filtered = append(filtered, line)
	}

	filteredContent := strings.Join(filtered, "\n")
	for pattern, replacement := range migrationDumpRemovePatterns {
		filteredContent = pattern.ReplaceAllString(filteredContent, replacement)
	}

	return fmt.Sprintf("BEGIN;\n\n%s\n\nCOMMIT;\n", strings.TrimSpace(filteredContent))
}

// removeAncestorsOf removes all migrations that are an ancestor of the given target version.
// This method returns the names of the files that were removed.
func removeAncestorsOf(database db.Database, ds *definition.Definitions, targetVersion int) ([]string, error) {
	allDefinitions := ds.All()

	allIDs := make([]int, 0, len(allDefinitions))
	for _, definition := range allDefinitions {
		allIDs = append(allIDs, definition.ID)
	}

	properDescendants, err := ds.Down(allIDs, []int{targetVersion})
	if err != nil {
		return nil, err
	}

	keep := make(map[int]struct{}, len(properDescendants))
	for _, definition := range properDescendants {
		keep[definition.ID] = struct{}{}
	}

	// Gather the set of filtered that are NOT a proper descendant of the given target version.
	// This will leave us with the ancestors of the target version (including itself).
	filtered := make([]string, 0, len(allDefinitions))
	for _, definition := range allDefinitions {
		if _, ok := keep[definition.ID]; !ok {
			filtered = append(filtered, strconv.Itoa(definition.ID))
		}
	}

	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return nil, err
	}

	for _, name := range filtered {
		if err := os.RemoveAll(filepath.Join(baseDir, name)); err != nil {
			return nil, err
		}
	}

	return filtered, nil
}
