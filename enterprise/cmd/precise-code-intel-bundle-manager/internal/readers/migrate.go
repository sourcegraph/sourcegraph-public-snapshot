package readers

import (
	"database/sql"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/cache"
)

func Migrate(bundleDir string, storeCache cache.StoreCache, db *sql.DB) error {
	paths, err := sqlitePaths(bundleDir)
	if err != nil {
		return err
	}
	if len(paths) == 0 {
		return nil
	}

	if err := migrateVersions(bundleDir, storeCache, paths); err != nil {
		return err
	}

	if err := migrateToPostgres(bundleDir, storeCache, db, paths); err != nil {
		return err
	}

	return nil
}
