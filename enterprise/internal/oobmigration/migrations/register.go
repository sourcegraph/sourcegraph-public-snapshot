package migrations

import (
	"context"
	"database/sql"

	workerCodeIntel "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	edb "github.com/sourcegraph/sourcegraph/enterprise/internal/database"
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
	codeIntelDB, err := workerCodeIntel.InitCodeIntelDatabase()
	if err != nil {
		return err
	}

	var codeInsightsDB edb.InsightsDB
	if internalInsights.IsEnabled() {
		codeInsightsDB, err = internalInsights.InitializeCodeInsightsDB("worker-oobmigrator")
		if err != nil {
			return err
		}
	}

	keyring := keyring.Default()

	return registerEnterpriseMigrations(outOfBandMigrationRunner, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: basestore.NewWithHandle(basestore.NewHandleWithDB(codeIntelDB, sql.TxOptions{})),
		insightsStore:  basestore.NewWithHandle(codeInsightsDB.Handle()),
		keyring:        &keyring,
	})
}

func RegisterEnterpriseMigrationsFromConfig(db database.DB, outOfBandMigrationRunner *oobmigration.Runner, conf conftypes.UnifiedQuerier) error {
	codeIntelDB, err := workerCodeIntel.InitCodeIntelDatabase() // TODO - get from config
	if err != nil {
		return err
	}

	var codeInsightsDB edb.InsightsDB
	if internalInsights.IsEnabled() {
		codeInsightsDB, err = internalInsights.InitializeCodeInsightsDB("migrator-oobmigrator") // TODO - get from config
		if err != nil {
			return err
		}
	}

	ctx := context.Background()
	keyring, err := keyring.NewRing(ctx, conf.SiteConfig().EncryptionKeys)
	if err != nil {
		return err
	}

	return registerEnterpriseMigrations(outOfBandMigrationRunner, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: basestore.NewWithHandle(basestore.NewHandleWithDB(codeIntelDB, sql.TxOptions{})),
		insightsStore:  basestore.NewWithHandle(codeInsightsDB.Handle()),
		keyring:        keyring,
	})
}

type dependencies struct {
	store          *basestore.Store
	codeIntelStore *basestore.Store
	insightsStore  *basestore.Store
	keyring        *keyring.Ring
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
