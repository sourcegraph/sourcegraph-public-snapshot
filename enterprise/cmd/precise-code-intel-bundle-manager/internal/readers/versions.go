package readers

import (
	"context"
	"os"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-bundle-manager/internal/paths"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/cache"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/persistence/sqlite/migrate"
)

// migrateVersions runs through each SQLite database on disk and opens a store instance which will perform
// any necessary migrations to transform it to the newest schema. Because this may have a non-negligible
// cost cost some intersection of migrations and database size, we try to pay this cost up-front instead
// of being paid on-demand when the database is opened within the query path. This method does not block
// the startup of the bundle manager as it does not change the correctness of the service.
func migrateVersions(bundleDir string, storeCache cache.StoreCache, bundleFilenames []string) error {
	version := migrate.CurrentSchemaVersion
	migrationMarkerFilename := paths.MigrationMarkerFilename(bundleDir, version)

	// If a file exists indicating the current schema version, then we've already run a full background
	// migration and can exit early. If the file doesn't exist, we'll run the migration and then write
	// to this file to indicate that we don't need to perform the migration again again in the future.
	if exists, err := paths.PathExists(migrationMarkerFilename); err != nil || exists {
		return err
	}

	log15.Info(
		"Migrating bundles to new version in background",
		"version", version,
		"numBundles", len(bundleFilenames),
	)

	for _, filename := range bundleFilenames {
		log15.Debug("Migrating bundle", "filename", filename)

		if err := storeCache.WithStore(context.Background(), filename, noopHandler); err != nil {
			log15.Error("Failed to migrate bundle", "err", err, "filename", filename)
		}
	}

	touchFile(migrationMarkerFilename)
	log15.Info("Finished bundle migration", "version", version)
	return nil
}

// touchFile ensures an empty file exists at the given path.
func touchFile(filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log15.Error("Failed to create migration marker", "err", err)
		return
	}
	if err := file.Close(); err != nil {
		log15.Error("Failed to create migration marker", "err", err)
		return
	}
}

func noopHandler(store persistence.Store) error {
	return nil
}
