package migrations

import (
	workerCodeIntel "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/iam"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
)

func RegisterEnterpriseMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	frontendStore, err := frontendStore(db)
	if err != nil {
		return err
	}

	codeIntelStore, err := codeIntelStore()
	if err != nil {
		return err
	}

	if err := insights.RegisterMigrations(db, outOfBandMigrationRunner); err != nil {
		return err
	}

	return migrations.RegisterAll(outOfBandMigrationRunner, []migrations.TaggedMigrator{
		iam.NewSubscriptionAccountNumberMigrator(frontendStore),
		iam.NewLicenseKeyFieldsMigrator(frontendStore),
		batches.NewSSHMigratorWithDB(frontendStore, keyring.Default().BatchChangesCredentialKey),
		codeintel.NewDiagnosticsCountMigrator(codeIntelStore, 1000),
		codeintel.NewDefinitionLocationsCountMigrator(codeIntelStore, 1000),
		codeintel.NewReferencesLocationsCountMigrator(codeIntelStore, 1000),
		codeintel.NewDocumentColumnSplitMigrator(codeIntelStore, 100),
		codeintel.NewAPIDocsSearchMigrator(),
	})
}

func frontendStore(db database.DB) (*basestore.Store, error) {
	return basestore.NewWithHandle(db.Handle()), nil
}

func codeIntelStore() (*basestore.Store, error) {
	lsifStore, err := workerCodeIntel.InitLSIFStore()
	if err != nil {
		return nil, err
	}

	return lsifStore.Store, err
}
