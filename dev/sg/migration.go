package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

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
func createNewMigration(databaseName, migrationName string) (up string, down string, _ error) {
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

	upPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.up.sql", lastMigrationIndex+1, migrationName))
	downPath := filepath.Join(baseDir, fmt.Sprintf("%d_%s.down.sql", lastMigrationIndex+1, migrationName))
	return upPath, downPath, writeMigrationFiles(upPath, downPath)
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
