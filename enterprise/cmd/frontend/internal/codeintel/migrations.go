package codeintel

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// registerMigrations registers all out-of-band migration instances that should run for
// the current version of Sourcegraph.
func registerMigrations(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	migrators := map[int]oobmigration.Migrator{
		migration.DiagnosticsCountMigrationID: migration.NewDiagnosticsCountMigrator(services.lsifStore, config.DiagnosticsCountMigrationBatchSize),
		migration.DefinitionsCountMigrationID: migration.NewLocationsCountMigrator(services.lsifStore, "lsif_data_definitions", config.DefinitionsCountMigrationBatchSize),
		migration.ReferencesCountMigrationID:  migration.NewLocationsCountMigrator(services.lsifStore, "lsif_data_references", config.ReferencesCountMigrationBatchSize),
	}

	for id, migrator := range migrators {
		if err := outOfBandMigrationRunner.Register(id, migrator, oobmigration.MigratorOptions{Interval: time.Second}); err != nil {
			return err
		}
	}

	return nil
}
