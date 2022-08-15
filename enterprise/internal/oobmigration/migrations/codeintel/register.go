package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	lsifmigrations "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/lsifstore/migration"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
)

type TaggedMigrator interface {
	oobmigration.Migrator

	ID() int
	Interval() time.Duration
}

// RegisterMigrations registers all code intel related out-of-band migration instances that should run for the current version of Sourcegraph.
func RegisterMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	if err := config.Validate(); err != nil {
		return err
	}

	lsifStore, err := codeintel.InitLSIFStore()
	if err != nil {
		return err
	}
	store := lsifStore.Store

	for _, m := range []TaggedMigrator{
		lsifmigrations.NewDiagnosticsCountMigrator(store, config.DiagnosticsCountMigrationBatchSize).(TaggedMigrator),
		lsifmigrations.NewDefinitionLocationsCountMigrator(store, config.DefinitionsCountMigrationBatchSize).(TaggedMigrator),
		lsifmigrations.NewReferencesLocationsCountMigrator(store, config.ReferencesCountMigrationBatchSize).(TaggedMigrator),
		lsifmigrations.NewDocumentColumnSplitMigrator(store, config.DocumentColumnSplitMigrationBatchSize).(TaggedMigrator),
		lsifmigrations.NewAPIDocsSearchMigrator(config.APIDocsSearchMigrationBatchSize).(TaggedMigrator),
	} {
		if err := outOfBandMigrationRunner.Register(m.ID(), m, oobmigration.MigratorOptions{Interval: m.Interval()}); err != nil {
			return err
		}
	}

	return nil
}
