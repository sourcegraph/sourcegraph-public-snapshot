package stitch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var migrationsArchive = &migrationArchives{currentVersion: "v5.3.0", m: map[string]migrationEntries{}}

func init() {
	err := migrationsArchive.load("foo")
	if err != nil {
		panic(err)
	}
}

// readMigrationDirectoryFilenames reads the names of the direct children of the given migration directory
// at the given git revision.
func readMigrationDirectoryFilenames(schemaName, dir, rev string) ([]string, error) {
	pathForSchemaAtRev, err := migrationPath(schemaName, rev)
	if err != nil {
		return nil, err
	}

	m, err := migrationsArchive.Get(rev)
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
func readMigrationFileContents(schemaName, dir, rev, path string) (string, error) {
	m, err := cachedArchiveContents(dir, rev)
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

var (
	revToPathTocontentsCacheMutex sync.RWMutex
	revToPathTocontentsCache      = map[string]map[string]string{}
)

// cachedArchiveContents memoizes archiveContents by git revision and schema name.
func cachedArchiveContents(dir, rev string) (map[string]string, error) {
	revToPathTocontentsCacheMutex.Lock()
	defer revToPathTocontentsCacheMutex.Unlock()

	m, ok := revToPathTocontentsCache[rev]
	if ok {
		return m, nil
	}

	m, err := archiveContents(dir, rev)
	if err != nil {
		return nil, err
	}

	revToPathTocontentsCache[rev] = m
	return m, nil
}

// archiveContents calls git archive with the given git revision and returns a map from
// file paths to file contents.
func archiveContents(dir, rev string) (map[string]string, error) {
	return migrationsArchive.Get(rev)
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

// tagRevToBranch attempts to determine the branch on which the given rev, assumed to be a tag of the
// form vX.Y.Z, belongs. This is used to support generation of stitched migrations after a branch cut
// but before the tagged commit is created.
func tagRevToBranch(rev string) (string, bool) {
	version, ok := oobmigration.NewVersionFromString(rev)
	if !ok {
		return "", false
	}

	return fmt.Sprintf("%d.%d", version.Major, version.Minor), true
}
