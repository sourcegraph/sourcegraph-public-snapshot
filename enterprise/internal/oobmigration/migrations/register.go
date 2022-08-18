package migrations

import (
	workerCodeIntel "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	internalInsights "github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/iam"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/insights"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
)

func RegisterEnterpriseMigrations(db database.DB, outOfBandMigrationRunner *oobmigration.Runner) error {
	codeIntelStore, err := codeIntelStore()
	if err != nil {
		return err
	}

	insightsStore, err := insightsStore()
	if err != nil {
		return err
	}

	return registerEnterpriseMigrations(outOfBandMigrationRunner, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: codeIntelStore,
		insightsStore:  insightsStore,
		keyring:        keyring.Default(),
	})
}

func RegisterEnterpriseMigrationsFromConfig(db database.DB, outOfBandMigrationRunner *oobmigration.Runner, conf conftypes.UnifiedQuerier) error {
	codeIntelStore, err := codeIntelStore() // TODO - get from config
	if err != nil {
		return err
	}

	insightsStore, err := insightsStore() // TODO - get from config
	if err != nil {
		return err
	}

	return registerEnterpriseMigrations(outOfBandMigrationRunner, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: codeIntelStore,
		insightsStore:  insightsStore,
		keyring:        keyring.Default(), // TODO - get from config
	})
}

type dependencies struct {
	store          *basestore.Store
	codeIntelStore *basestore.Store
	insightsStore  *basestore.Store
	keyring        keyring.Ring
}

func registerEnterpriseMigrations(outOfBandMigrationRunner *oobmigration.Runner, deps dependencies) error {
	return migrations.RegisterAll(outOfBandMigrationRunner, []migrations.TaggedMigrator{
		iam.NewSubscriptionAccountNumberMigrator(deps.store, 500),
		iam.NewLicenseKeyFieldsMigrator(deps.store, 500),
		batches.NewSSHMigratorWithDB(deps.store, deps.keyring.BatchChangesCredentialKey, 5),
		codeintel.NewDiagnosticsCountMigrator(deps.codeIntelStore, 1000),
		codeintel.NewDefinitionLocationsCountMigrator(deps.codeIntelStore, 1000),
		codeintel.NewReferencesLocationsCountMigrator(deps.codeIntelStore, 1000),
		codeintel.NewDocumentColumnSplitMigrator(deps.codeIntelStore, 100),
		codeintel.NewAPIDocsSearchMigrator(),
		insights.NewMigrator(deps.store, deps.insightsStore),
	})
}

func codeIntelStore() (*basestore.Store, error) {
	lsifStore, err := workerCodeIntel.InitLSIFStore()
	if err != nil {
		return nil, err
	}

	return lsifStore.Store, err
}

func insightsStore() (*basestore.Store, error) {
	if !internalInsights.IsEnabled() {
		return nil, nil
	}

	db, err := internalInsights.InitializeCodeInsightsDB("worker-oobmigrator")
	if err != nil {
		return nil, err
	}

	return basestore.NewWithHandle(db.Handle()), nil
}
