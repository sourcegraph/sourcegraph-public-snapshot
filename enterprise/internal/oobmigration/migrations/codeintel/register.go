package codeintel

import (
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
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
		NewDiagnosticsCountMigrator(store, config.DiagnosticsCountMigrationBatchSize),
		NewDefinitionLocationsCountMigrator(store, config.DefinitionsCountMigrationBatchSize),
		NewReferencesLocationsCountMigrator(store, config.ReferencesCountMigrationBatchSize),
		NewDocumentColumnSplitMigrator(store, config.DocumentColumnSplitMigrationBatchSize),
		NewAPIDocsSearchMigrator(),
	} {
		if err := outOfBandMigrationRunner.Register(m.ID(), m, oobmigration.MigratorOptions{Interval: m.Interval()}); err != nil {
			return err
		}
	}

	return nil
}
