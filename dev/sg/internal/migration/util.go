package migration

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/db"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
	"github.com/sourcegraph/sourcegraph/internal/database/migration/definition"
)

// readDefinitions returns definitions from the given database object.
func readDefinitions(database db.Database) (*definition.Definitions, error) {
	fs, err := database.FS()
	if err != nil {
		return nil, err
	}

	return definition.ReadDefinitions(fs)
}

// makeMigrationFilenames makes a pair of (absolute) paths to migration files with the given migration index.
func makeMigrationFilenames(database db.Database, migrationIndex int) (up, down, metadata string, _ error) {
	baseDir, err := migrationDirectoryForDatabase(database)
	if err != nil {
		return "", "", "", err
	}

	upPath := filepath.Join(baseDir, fmt.Sprintf("%d/up.sql", migrationIndex))
	downPath := filepath.Join(baseDir, fmt.Sprintf("%d/down.sql", migrationIndex))
	metadataPath := filepath.Join(baseDir, fmt.Sprintf("%d/metadata.yaml", migrationIndex))
	return upPath, downPath, metadataPath, nil
}

// migrationDirectoryForDatabase returns the directory where migration files are stored for the
// given database.
func migrationDirectoryForDatabase(database db.Database) (string, error) {
	repoRoot, err := root.RepositoryRoot()
	if err != nil {
		return "", err
	}

	return filepath.Join(repoRoot, "migrations", database.Name), nil
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
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		if err := os.WriteFile(path, []byte(contents), os.FileMode(0644)); err != nil {
			return err
		}
	}

	return nil
}
