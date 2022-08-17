package migrations

import (
	workercodeintel "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/codeintel"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
)

func RegisterEnterpriseMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	lsifStore, err := workercodeintel.InitLSIFStore()
	if err != nil {
		return err
	}
	store := lsifStore.Store

	if err := insights.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	return migrations.RegisterAll(outOfBandMigrationRunner, []migrations.TaggedMigrator{
		NewSubscriptionAccountNumberMigrator(db),
		NewLicenseKeyFieldsMigrator(db),
		batches.NewSSHMigratorWithDB(db, keyring.Default().BatchChangesCredentialKey),
		codeintel.NewDiagnosticsCountMigrator(store, 1000),
		codeintel.NewDefinitionLocationsCountMigrator(store, 1000),
		codeintel.NewReferencesLocationsCountMigrator(store, 1000),
		codeintel.NewDocumentColumnSplitMigrator(store, 100),
		codeintel.NewAPIDocsSearchMigrator(),
	})
}
