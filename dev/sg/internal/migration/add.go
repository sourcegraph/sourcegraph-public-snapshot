package migration

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
)

const metadataFileTemplate = `name: %s
parents: [%s]
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

// Add creates a new directory with stub migration files in the given schema and returns the
// names of the newly created files. If there was an error, the filesystem is rolled-back.
func Add(database db.Database, migrationName string) (up, down, metadata string, _ error) {
	definitions, err := readDefinitions(database)
	if err != nil {
		return "", "", "", err
	}

	leaves := definitions.Leaves()
	parents := make([]int, 0, len(leaves))
	for _, leaf := range leaves {
		parents = append(parents, leaf.ID)
	}

	id := int(time.Now().UTC().Unix())

	upPath, downPath, metadataPath, err := makeMigrationFilenames(database, id)
	if err != nil {
		return "", "", "", err
	}

	contents := map[string]string{
		upPath:       upMigrationFileTemplate,
		downPath:     downMigrationFileTemplate,
		metadataPath: fmt.Sprintf(metadataFileTemplate, migrationName, strings.Join(intsToStrings(parents), ", ")),
	}
	if err := writeMigrationFiles(contents); err != nil {
		return "", "", "", err
	}

	return upPath, downPath, metadataPath, nil
}

func intsToStrings(ints []int) []string {
	strs := make([]string, 0, len(ints))
	for _, value := range ints {
		strs = append(strs, strconv.Itoa(value))
	}

	return strs
}
