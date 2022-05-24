package migrations

import (
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	dbmigrations "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore/migration"
	lsifmigrations "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore/migration"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

// RegisterMigrations registers all code intel related out-of-band migration instances that should run for the current version of Sourcegraph.
func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	if err := config.Validate(); err != nil {
		return err
	}

	dbStore, err := codeintel.InitDBStore()
	if err != nil {
		return err
	}

	lsifStore, err := codeintel.InitLSIFStore()
	if err != nil {
		return err
	}

	gitserverClient, err := codeintel.InitGitserverClient()
	if err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.DiagnosticsCountMigrationID, // 1
		lsifmigrations.NewDiagnosticsCountMigrator(lsifStore, config.DiagnosticsCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DiagnosticsCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.DefinitionsCountMigrationID, // 4
		lsifmigrations.NewLocationsCountMigrator(lsifStore, "lsif_data_definitions", config.DefinitionsCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DefinitionsCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.ReferencesCountMigrationID, // 5
		lsifmigrations.NewLocationsCountMigrator(lsifStore, "lsif_data_references", config.ReferencesCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.ReferencesCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.DocumentColumnSplitMigrationID, // 7
		lsifmigrations.NewDocumentColumnSplitMigrator(lsifStore, config.DocumentColumnSplitMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.DocumentColumnSplitMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		lsifmigrations.APIDocsSearchMigrationID, // 12
		lsifmigrations.NewAPIDocsSearchMigrator(config.APIDocsSearchMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.APIDocsSearchMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		dbmigrations.CommittedAtMigrationID, // 8
		dbmigrations.NewCommittedAtMigrator(dbStore, gitserverClient, config.CommittedAtMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.CommittedAtMigrationBatchInterval},
	); err != nil {
		return err
	}

	if err := outOfBandMigrationRunner.Register(
		dbmigrations.ReferenceCountMigrationID, // 11
		dbmigrations.NewReferenceCountMigrator(dbStore, config.ReferenceCountMigrationBatchSize),
		oobmigration.MigratorOptions{Interval: config.ReferenceCountMigrationBatchInterval},
	); err != nil {
		return err
	}

	return nil
}
