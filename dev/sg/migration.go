package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
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

// createNewMigration creates a new up/down migration file pair for the given database.
func createNewMigration(databaseName, migrationName string) (up string, down string, err error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", "", err
	}

	entries, err := os.ReadDir(filepath.Join(repoRoot, "migrations", databaseName))
	if err != nil {
		return "", "", err
	}

	indices := make([]int, 0, len(entries))
	for _, entry := range entries {
		if value, err := strconv.Atoi(strings.Split(entry.Name(), "_")[0]); err == nil {
			indices = append(indices, value)
		}
	}

	upPath := filepath.Join(repoRoot, "migrations", databaseName, fmt.Sprintf("%d_%s.up.sql", indices[len(indices)-1]+1, migrationName))
	downPath := filepath.Join(repoRoot, "migrations", databaseName, fmt.Sprintf("%d_%s.down.sql", indices[len(indices)-1]+1, migrationName))
	paths := []string{upPath, downPath}

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
			return "", "", err
		}
	}

	return upPath, downPath, nil
}
