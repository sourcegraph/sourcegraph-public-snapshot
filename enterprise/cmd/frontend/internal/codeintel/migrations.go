package codeintel

import (
	"context"

	dbmigrations "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore/migration"
	lsifmigrations "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/lsifstore/migration"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// registerMigrations registers all out-of-band migration instances that should run for
// the current version of Sourcegraph.
func registerMigrations(ctx context.Context, db dbutil.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.DiagnosticsCountMigrationID, // 1
		lsifmigrations.NewDiagnosticsCountMigrator(services.lsifStore, config.DiagnosticsCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DiagnosticsCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.DefinitionsCountMigrationID, // 4
		lsifmigrations.NewLocationsCountMigrator(services.lsifStore, "lsif_data_definitions", config.DefinitionsCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DefinitionsCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.ReferencesCountMigrationID, // 5
		lsifmigrations.NewLocationsCountMigrator(services.lsifStore, "lsif_data_references", config.ReferencesCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.ReferencesCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.DocumentColumnSplitMigrationID, // 7
		lsifmigrations.NewDocumentColumnSplitMigrator(services.lsifStore, config.DocumentColumnSplitMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DocumentColumnSplitMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		dbmigrations.CommittedAtMigrationID, // 8
		dbmigrations.NewCommittedAtMigrator(services.dbStore, services.gitserverClient, config.CommittedAtMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.CommittedAtMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		dbmigrations.ReferenceCountMigrationID, // 11
		dbmigrations.NewReferenceCountMigrator(services.dbStore, config.ReferenceCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.ReferenceCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	return nil
}
