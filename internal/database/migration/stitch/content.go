package stitch

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// readMigrationDirectoryFilenames reads the names of the direct children of the given migration directory
// at the given git revision.
func readMigrationDirectoryFilenames(ma MigrationsReader, schemaName, rev string) ([]string, error) {
	pathForSchemaAtRev, err := migrationPath(schemaName, rev)
	if err != nil {
		return nil, err
	}

	m, err := ma.Get(rev)
	if err != nil {
		return nil, err
	}

	entries := []string{}
	for path := range m {
		if strings.HasPrefix(path, pathForSchemaAtRev) && path != pathForSchemaAtRev {
			entries = append(entries, strings.Replace(path, pathForSchemaAtRev+"/", "", 1))
		}
	}
	return entries, nil
}

// readMigrationFileContents reads the contents of the migration at given path at the given git revision.
func readMigrationFileContents(ma MigrationsReader, schemaName, rev, path string) (string, error) {
	m, err := ma.Get(rev)
	if err != nil {
		return "", err
	}

	pathForSchemaAtRev, err := migrationPath(schemaName, rev)
	if err != nil {
		return "", err
	}
	if v, ok := m[filepath.Join(pathForSchemaAtRev, path)]; ok {
		return v, nil
	}

	return "", os.ErrNotExist
}

func migrationPath(schemaName, rev string) (string, error) {
	revVersion, ok := oobmigration.NewVersionFromString(rev)
	if !ok {
		return "", errors.Newf("illegal rev %q", rev)
	}
	if oobmigration.CompareVersions(revVersion, oobmigration.NewVersion(3, 21)) == oobmigration.VersionOrderBefore {
		if schemaName == "frontend" {
			// Return the root directory if we're looking for the frontend schema
			// at or before 3.20. This was the only schema in existence then.
			return "migrations", nil
		}
	}

	return filepath.Join("migrations", schemaName), nil
}
