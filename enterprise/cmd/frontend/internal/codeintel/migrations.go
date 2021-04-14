package codeintel

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// registerMigrations registers all out-of-band migration instances that should run for
// the current version of Sourcegraph.
func registerMigrations(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	if err := outOfBandMigrationRunner.Register(
		migration.DiagnosticsCountMigrationID, // 1
		migration.NewDiagnosticsCountMigrator(services.lsifStore, config.DiagnosticsCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DiagnosticsCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		migration.DefinitionsCountMigrationID, // 4
		migration.NewLocationsCountMigrator(services.lsifStore, "lsif_data_definitions", config.DefinitionsCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DefinitionsCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		migration.ReferencesCountMigrationID, // 5
		migration.NewLocationsCountMigrator(services.lsifStore, "lsif_data_references", config.ReferencesCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.ReferencesCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		migration.DocumentColumnSplitMigrationID, // 7
		migration.NewDocumentColumnSplitMigrator(services.lsifStore, config.DocumentColumnSplitMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DocumentColumnSplitMigrationBatchInterval},
	); err != nil {
		return err
	}

	return nil
}
