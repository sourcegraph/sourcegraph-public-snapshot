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
)

var (
	databaseNames = []string{
		"frontend",
		"codeintel",
		"codeinsights",
	}

	defaultDatabaseName = databaseNames[0]
)

// TODO - document
func isValidDatabaseName(name string) bool {
	for _, candidate := range databaseNames {
		if candidate == name {
			return true
		}
	}

	return false
}

// TODO - document
func migrationsTableForDatabase(databaseName string) string {
	if databaseName == defaultDatabaseName {
		return "schema_migrations"
	}

	return fmt.Sprintf("%s_schema_migrations", databaseName)
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

	if err := writeMigrationFiles(upPath, downPath); err != nil {
		return "", "", err
	}

	return upPath, downPath, nil
}

// TODO - document
func makeMigrationFilenames(databaseName string, migrationIndex int, migrationName string) (up string, down string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(databaseName)
	if err != nil {
		return "", "", err
	}

	upPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.up.sql", migrationIndex, migrationName))
	downPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.down.sql", migrationIndex, migrationName))
	return upPath, downPath, nil
}

// removeMigrationFilesBefore removes migration files for the given database falling on or
// before the given migration index. This method returns the names of the files that were
// removed.
func removeMigrationFilesBefore(databaseName string, targetIndex int) ([]string, error) {
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

// migrationDirectoryForDatabase returns the directory where migration files are stored for the
// given database.
func migrationDirectoryForDatabase(databaseName string) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(repoRoot, "migrations", databaseName), nil
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

const (
	squasherContainerName        = "squasher"
	squasherContainerExposedPort = 5433
)

const squashedDownMigrationTemplate = `
DROP SCHEMA IF EXISTS public CASCADE;
CREATE SCHEMA public;

CREATE TABLE IF NOT EXISTS %s (
	version bigint NOT NULL PRIMARY KEY,
	dirty boolean NOT NULL
);
`

// TODO - document
func generateSquashedMigrations(databaseName string, migrationIndex int) (up, down string, err error) {
	baseDir, err := migrationDirectoryForDatabase(databaseName)
	if err != nil {
		return "", "", err
	}
	migrationsTable := migrationsTableForDatabase(databaseName)

	runArgs := []string{
		"run",
		"--rm", "-d",
		"--name", squasherContainerName,
		"-p", fmt.Sprintf("%d:5432", squasherContainerExposedPort),
		"-e", "POSTGRES_HOST_AUTH_METHOD=trust",
		"postgres:12.6",
	}
	if _, err := runDockerCmd(runArgs...); err != nil {
		return "", "", err
	}
	defer func() {
		killArgs := []string{
			"kill",
			squasherContainerName,
		}
		if _, killErr := runDockerCmd(killArgs...); killErr != nil {
			err = multierror.Append(err, fmt.Errorf("failed to stop docker container: %s", killErr))
		}
	}()

	// TODO - check health instead
	time.Sleep(5 * time.Second)

	fmt.Printf("CREATING DB\n")

	execArgs := []string{
		"exec",
		"-u", "postgres",
		squasherContainerName,
		"createdb", databaseName,
	}
	if _, err := runDockerCmd(execArgs...); err != nil {
		return "", "", err
	}

	fmt.Printf("MIGRATING\n")

	if _, err := runCommandInRoot(exec.Command(
		"migrate",
		"-database", fmt.Sprintf(
			"postgres://postgres@127.0.0.1:%d/%s?sslmode=disable&x-migrations-table=%s",
			squasherContainerExposedPort,
			databaseName,
			migrationsTable,
		),
		"-path", baseDir,
		"goto", strconv.FormatInt(int64(migrationIndex), 10),
	)); err != nil {
		return "", "", err
	}

	fmt.Printf("DUMPING\n")

	cmd := exec.Command(
		"pg_dump",
		"--exclude-table='*schema_migrations'",
		"--no-owner",
		"--schema-only",
	)
	cmd.Env = []string{
		"PGHOST=127.0.0.1",
		fmt.Sprintf("PGPORT=%d", squasherContainerExposedPort),
		fmt.Sprintf("PGDATABASE=%s", databaseName),
		"PGUSER=postgres",
	}
	pgDumpOutput, err := runCommandInRoot(cmd)
	if err != nil {
		return "", "", err
	}

	return sanitizePgDumpOutput(pgDumpOutput), fmt.Sprintf(squashedDownMigrationTemplate, migrationsTable), nil
}

var (
	migrationDumpRemovePrefixes = []string{
		"--",                            // remove comments
		"SET ",                          // remove settings header
		"SELECT pg_catalog.set_config.", // remove settings header
		"DROP EXTENSION ",               // do not drop extensions if they already exist
	}

	migrationDumpRemovePatterns = map[*regexp.Regexp]string{
		regexp.MustCompile(`\bpublic\.`):             "",
		regexp.MustCompile(`\bWITH SCHEMA public\b`): "",
		regexp.MustCompile(`\n{3,}`):                 "\n\n",
	}
)

// TODO - document
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
