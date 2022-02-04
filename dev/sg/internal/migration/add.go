package migration

import (
	"fmt"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const metadataFileTemplate = `name: %s
parent: %d
`

const upMigrationFileTemplate = `BEGIN;

-- Perform migration here.
--
-- See /migrations/README.md. Highlights:
--  * Make migrations idempotent (use IF EXISTS)
--  * Make migrations backwards-compatible (old readers/writers must continue to work)
--  * Wrap your changes in a transaction
--  * If you are using CREATE INDEX CONCURRENTLY, then make sure that only one statement
--    is defined per file, and that each such statement is NOT wrapped in a transaction.
--    Each such migration must also declare "createIndexConcurrently: true" in their
--    associated metadata.yaml file.

COMMIT;
`

const downMigrationFileTemplate = `BEGIN;

-- Undo the changes made in the up migration

COMMIT;
`

// Add creates a new up/down migration file pair for the given database and
// returns the names of the new files. If there was an error, the filesystem should remain
// unmodified.
func Add(database db.Database, migrationName string) (up, down, metadata string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return "", "", "", err
	}

	// TODO: We can probably convert to migrations and use getMaxMigrationID
	names, err := readFilenamesNamesInDirectory(baseDir)
	if err != nil {
		return "", "", "", err
	}

	lastMigrationIndex, ok := parseLastMigrationIndex(names)
	if !ok {
		return "", "", "", errors.New("no previous migrations exist")
	}

	upPath, downPath, metadataPath, err := makeMigrationFilenames(database, lastMigrationIndex+1)
	if err != nil {
		return "", "", "", err
	}

	contents := map[string]string{
		upPath:       upMigrationFileTemplate,
		downPath:     downMigrationFileTemplate,
		metadataPath: fmt.Sprintf(metadataFileTemplate, migrationName, lastMigrationIndex),
	}
	if err := writeMigrationFiles(contents); err != nil {
		return "", "", "", err
	}

	return upPath, downPath, metadataPath, nil
}
