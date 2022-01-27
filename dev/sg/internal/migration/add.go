package migration

import (
	"fmt"
	"os"

	"github.com/cockroachdb/errors"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
)

const upMigrationFileTemplate = `-- +++
-- parent: %d
-- +++

BEGIN;

-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * Wrap your changes in a transaction. Note that CREATE INDEX CONCURRENTLY is an exception
--    and cannot be performed in a transaction. For such migrations, ensure that only one
--    statement is defined per migration to prevent query transactions from starting implicitly.

COMMIT;
`

const downMigrationFileTemplate = `BEGIN;

-- Undo the changes made in the up migration

COMMIT;
`

// Add creates a new up/down migration file pair for the given database and
// returns the names of the new files. If there was an error, the filesystem should remain
// unmodified.
func Add(database db.Database, migrationName string) (up, down string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return "", "", err
	}

	// TODO: We can probably convert to migrations and use getMaxMigrationID
	names, err := readFilenamesNamesInDirectory(baseDir)
	if err != nil {
		return "", "", err
	}

	lastMigrationIndex, ok := parseLastMigrationIndex(names)
	if !ok {
		return "", "", errors.New("no previous migrations exist")
	}

	upPath, downPath, err := makeMigrationFilenames(database, lastMigrationIndex+1, migrationName)
	if err != nil {
		return "", "", err
	}

	contents := map[string]string{
		upPath:   fmt.Sprintf(upMigrationFileTemplate, lastMigrationIndex),
		downPath: downMigrationFileTemplate,
	}

	if err := writeMigrationFiles(contents); err != nil {
		return "", "", err
	}

	return upPath, downPath, nil
}

// writeMigrationFiles writes the contents of migrationFileTemplate to the given filepaths.
func writeMigrationFiles(contents map[string]string) (err error) {
	defer func() {
		if err != nil {
			for path := range contents {
				// undo any changes to the fs on error
				_ = os.Remove(path)
			}
		}
	}()

	for path, contents := range contents {
		if err := os.WriteFile(path, []byte(contents), os.FileMode(0644)); err != nil {
			return err
		}
	}

	return nil
}
