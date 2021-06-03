package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

var (
	databaseNames = []string{
		"frontend",
		"codeintel",
		"codeinsights",
	}

	defaultDatabaseName = databaseNames[0]

	dataTables = map[string][]string{
		"frontend": {"out_of_band_migrations"},
	}
)

func isValidDatabaseName(name string) bool {
	for _, candidate := range databaseNames {
		if candidate == name {
			return true
		}
	}

	return false
}

// createNewMigration creates a new up/down migration file pair for the given database and
// returns the names of the new files. If there was an error, the filesystem should remain
// unmodified.
func createNewMigration(databaseName, migrationName string) (up, down string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(databaseName)
	if err != nil {
		return "", "", err
	}

	names, err := readFilenamesNamesInDirectory(baseDir)
	if err != nil {
		return "", "", err
	}

	lastMigrationIndex, ok := parseLastMigrationIndex(names)
	if !ok {
		return "", "", errors.New("no previous migrations exist")
	}

	upPath, downPath, err := makeMigrationFilenames(databaseName, lastMigrationIndex+1, migrationName)
	if err != nil {
		return "", "", err
	}

	if err := writeMigrationFiles(upPath, downPath); err != nil {
		return "", "", err
	}

	return upPath, downPath, nil
}

// generateSquashedMigrations generates the content of a migration file pair that contains the contents
// of a database up to a given migration index. This function will launch a daemon Postgres container,
// migrate a fresh database up to the given migration index, then dump and sanitize the contents.
func generateSquashedMigrations(databaseName string, migrationIndex int) (up, down string, err error) {
	postgresDSN := fmt.Sprintf(
		"postgres://postgres@127.0.0.1:%d/%s?sslmode=disable",
		squasherContainerExposedPort,
		databaseName,
	)

	teardown, err := runPostgresContainer(databaseName)
	if err != nil {
		return "", "", err
	}
	defer func() {
		err = teardown(err)
	}()

	if err := runMigrationsGoto(databaseName, migrationIndex, postgresDSN); err != nil {
		return "", "", err
	}

	upMigration, err := generateSquashedUpMigration(databaseName, postgresDSN)
	if err != nil {
		return "", "", err
	}

	downMigration, err := generateSquashedDownMigration(databaseName)
	if err != nil {
		return "", "", err
	}

	return upMigration, downMigration, nil
}

// removeMigrationFilesUpToIndex removes migration files for the given database falling on
// or before the given migration index. This method returns the names of the files that were
// removed.
func removeMigrationFilesUpToIndex(databaseName string, targetIndex int) ([]string, error) {
	baseDir, err := migrationDirectoryForDatabase(databaseName)
	if err != nil {
		return nil, err
	}

	names, err := readFilenamesNamesInDirectory(baseDir)
	if err != nil {
		return nil, err
	}

	filtered := names[:0]
	for _, name := range names {
		index, ok := parseMigrationIndex(name)
		if !ok {
			continue
		}

		if index <= targetIndex {
			filtered = append(filtered, name)
		}
	}

	for _, name := range filtered {
		if err := os.Remove(filepath.Join(baseDir, name)); err != nil {
			return nil, err
		}
	}

	return filtered, nil
}

// lastMigrationIndexAtCommit returns the index of the last migration for the given database
// name available at the given commit. This function returns a false-valued flag if no migrations
// exist at the given commit.
func lastMigrationIndexAtCommit(databaseName, commit string) (int, bool, error) {
	migrationsDir := filepath.Join("migrations", databaseName)

	output, err := runGitCmd("ls-tree", "-r", "--name-only", commit, migrationsDir)
	if err != nil {
		return 0, false, err
	}

	lastMigrationIndex, ok := parseLastMigrationIndex(strings.Split(string(output), "\n"))
	return lastMigrationIndex, ok, nil
}

// migrationsTableForDatabase returns the name of the migration table for the given database.
func migrationsTableForDatabase(databaseName string) string {
	if databaseName == defaultDatabaseName {
		return "schema_migrations"
	}

	return fmt.Sprintf("%s_schema_migrations", databaseName)
}

// migrationDirectoryForDatabase returns the directory where migration files are stored for the
// given database.
func migrationDirectoryForDatabase(databaseName string) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(repoRoot, "migrations", databaseName), nil
}

// makeMigrationFilenames makes a pair of (absolute) paths to migration files with the
// given  migration index and name.
func makeMigrationFilenames(databaseName string, migrationIndex int, migrationName string) (up string, down string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(databaseName)
	if err != nil {
		return "", "", err
	}

	upPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.up.sql", migrationIndex, migrationName))
	downPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.down.sql", migrationIndex, migrationName))
	return upPath, downPath, nil
}

// parseMigrationIndex parse a filename and returns the migration index if the filename
// looks like a migration. Each migration filename has the form {unique_id}_{name}.{dir}.sql.
// This function returns a false-valued flag on failure. Leading directories are stripped
// from the input, so a basename or a full path can be supplied.
func parseMigrationIndex(name string) (int, bool) {
	index, err := strconv.Atoi(strings.Split(filepath.Base(name), "_")[0])
	if err != nil {
		return 0, false
	}

	return index, true
}

// parseLastMigrationIndex parses a list of filenames and returns the highest migration
// index available.
func parseLastMigrationIndex(names []string) (int, bool) {
	indices := make([]int, 0, len(names))
	for _, name := range names {
		if index, ok := parseMigrationIndex(name); ok {
			indices = append(indices, index)
		}
	}
	sort.Ints(indices)

	if len(indices) == 0 {
		return 0, false
	}

	return indices[len(indices)-1], true
}

const migrationFileTemplate = `
BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

COMMIT;
`

// writeMigrationFiles writes the contents of migrationFileTemplate to the given filepaths.
func writeMigrationFiles(paths ...string) (err error) {
	defer func() {
		if err != nil {
			for _, path := range paths {
				// undo any changes to the fs on error
				_ = os.Remove(path)
			}
		}
	}()

	for _, path := range paths {
		if err := ioutil.WriteFile(path, []byte(migrationFileTemplate), os.ModePerm); err != nil {
			return err
		}
	}

	return nil
}

const (
	squasherContainerName        = "squasher"
	squasherContainerExposedPort = 5433
)

// runPostgresContainer runs a postgres:12.6 daemon with an empty db with the given name.
// This method returns a teardown function that filters the error value of the calling
// function, as well as any immediate synchronous error.
func runPostgresContainer(databaseName string) (_ func(err error) error, err error) {
	pending := out.Pending(output.Line("", output.StylePending, "Starting PostgreSQL 12 in a container..."))
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
		if _, killErr := runDockerCmd(killArgs...); killErr != nil {
			err = multierror.Append(err, fmt.Errorf("failed to stop docker container: %s", killErr))
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
	if _, err := runDockerCmd(runArgs...); err != nil {
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
	if _, err := runDockerCmd(execArgs...); err != nil {
		return nil, teardown(err)
	}

	return teardown, nil
}

// runMigrationsGoto runs the `migrate` utility to migrate up or down to the given
// migration index.
func runMigrationsGoto(databaseName string, migrationIndex int, postgresDSN string) (err error) {
	pending := out.Pending(output.Line("", output.StylePending, "Migrating PostgreSQL schema..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Migrated PostgreSQL schema"))
		} else {
			pending.Destroy()
		}
	}()

	_, err = runMigrate(
		databaseName,
		postgresDSN+fmt.Sprintf("&x-migrations-table=%s", migrationsTableForDatabase(databaseName)),
		"goto", strconv.FormatInt(int64(migrationIndex), 10),
	)
	return err
}

// runMigrationsUp runs the `migrate` utility to migrate up the given number of steps.
// If n is nil then all migrations are ran.
func runMigrationsUp(databaseName string, n *int) (string, error) {
	args := []string{"up"}
	if n != nil {
		args = append(args, strconv.Itoa(*n))
	}

	return runMigrate(databaseName, makePostgresDSN(databaseName), args...)
}

// runMigrationsDown runs the `migrate` utility to migrate up the given number of steps.
func runMigrationsDown(databaseName string, n int) (string, error) {
	return runMigrate(databaseName, makePostgresDSN(databaseName), "down", strconv.Itoa(n))
}

// runMigrate runs the migrate utility with the given arguments.
// TODO - replace with our db utilities
func runMigrate(databaseName, postgresDSN string, args ...string) (string, error) {
	baseDir, err := migrationDirectoryForDatabase(databaseName)
	if err != nil {
		return "", err
	}

	return runCommandInRoot(exec.Command("migrate", append([]string{"-database", postgresDSN, "-path", baseDir}, args...)...))
}

// makePostgresDSN returns a PostresDSN for the given database. For all databases except
// the default, the PG* environment variables are prefixed with the database name. The
// resulting address depends on the environment.
func makePostgresDSN(databaseName string) string {
	var prefix string
	if databaseName != defaultDatabaseName {
		prefix = strings.ToUpper(databaseName) + "_"
	}

	var port string
	if value := os.Getenv(fmt.Sprintf("%sPGPORT", prefix)); value != "" {
		port = ":" + value
	}

	return fmt.Sprintf(
		"postgres://%s%s/%s?x-migrations-table=%s",
		os.Getenv(fmt.Sprintf("%sPGHOST", prefix)),
		port,
		os.Getenv(fmt.Sprintf("%sPGDATABASE", prefix)),
		migrationsTableForDatabase(databaseName),
	)
}

// generateSquashedUpMigration returns the contents of an up migration file containing the
// current contents of the given database.
func generateSquashedUpMigration(databaseName, postgresDSN string) (_ string, err error) {
	pending := out.Pending(output.Line("", output.StylePending, "Dumping current database..."))
	defer func() {
		if err == nil {
			pending.Complete(output.Line(output.EmojiSuccess, output.StyleSuccess, "Dumped current database"))
		} else {
			pending.Destroy()
		}
	}()

	pgDump := func(args ...string) (string, error) {
		cmd := exec.Command("pg_dump", append([]string{postgresDSN}, args...)...)
		cmd.Env = []string{}
		return runCommandInRoot(cmd)
	}

	pgDumpOutput, err := pgDump("--schema-only", "--no-owner", "--exclude-table", "*schema_migrations")
	if err != nil {
		return "", err
	}

	for _, table := range dataTables[databaseName] {
		dataOutput, err := pgDump("--data-only", "--inserts", "--table", table)
		if err != nil {
			return "", err
		}

		pgDumpOutput += dataOutput
	}

	return sanitizePgDumpOutput(pgDumpOutput), nil
}

const squashedDownMigrationTemplate = `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

CREATE TABLE IF NOT EXISTS %s (
	version bigint NOT NULL PRIMARY KEY,
	dirty boolean NOT NULL
);
`

// generateSquashedDownMigration returns the contents of a down migration file containing the
// canned down migration for this database.
func generateSquashedDownMigration(databaseName string) (string, error) {
	return strings.TrimSpace(fmt.Sprintf(squashedDownMigrationTemplate, migrationsTableForDatabase(databaseName))) + "\n", nil
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

// readFilenamesNamesInDirectory returns a list of names in the given directory.
func readFilenamesNamesInDirectory(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}

	return names, nil
}
