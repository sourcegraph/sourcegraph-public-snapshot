package migrations

import (
	"context"
	"database/sql"

	workerCodeIntel "github.com/sourcegraph/sourcegraph/cmd/worker/shared/init/codeintel"
	internalInsights "github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/iam"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/insights"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func RegisterEnterpriseMigrations(ctx context.Context, db database.DB, runner *oobmigration.Runner) error {
	codeIntelDB, err := workerCodeIntel.InitCodeIntelDatabase()
	if err != nil {
		return err
	}

	var insightsStore *basestore.Store
	if internalInsights.IsEnabled() {
		codeInsightsDB, err := internalInsights.InitializeCodeInsightsDB("worker-oobmigrator")
		if err != nil {
			return err
		}

		insightsStore = basestore.NewWithHandle(codeInsightsDB.Handle())
	}

	keyring := keyring.Default()

	return registerEnterpriseMigrations(runner, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: basestore.NewWithHandle(basestore.NewHandleWithDB(codeIntelDB, sql.TxOptions{})),
		insightsStore:  insightsStore,
		keyring:        &keyring,
	})
}

func RegisterEnterpriseMigrationsFromConfig(ctx context.Context, db database.DB, runner *oobmigration.Runner, conf conftypes.UnifiedQuerier) error {
	codeIntelDB, err := connections.EnsureNewCodeIntelDB(
		conf.ServiceConnections().CodeIntelPostgresDSN,
		"migrator",
		&observation.TestContext,
	)
	if err != nil {
		return errors.Errorf("failed to connect to codeintel database: %s", err)
	}

	var insightsStore *basestore.Store
	if internalInsights.IsEnabled() {
		codeInsightsDB, err := connections.EnsureNewCodeInsightsDB(
			conf.ServiceConnections().CodeInsightsDSN,
			"migrator",
			&observation.TestContext,
		)
		if err != nil {
			return errors.Errorf("failed to connect to codeintel database: %s", err)
		}

		insightsStore = basestore.NewWithHandle(basestore.NewHandleWithDB(codeInsightsDB, sql.TxOptions{}))
	}

	keyring, err := keyring.NewRing(ctx, conf.SiteConfig().EncryptionKeys)
	if err != nil {
		return err
	}

	return registerEnterpriseMigrations(runner, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: basestore.NewWithHandle(basestore.NewHandleWithDB(codeIntelDB, sql.TxOptions{})),
		insightsStore:  insightsStore,
		keyring:        keyring,
	})
}

type dependencies struct {
	store          *basestore.Store
	codeIntelStore *basestore.Store
	insightsStore  *basestore.Store
	keyring        *keyring.Ring
}

func registerEnterpriseMigrations(runner *oobmigration.Runner, deps dependencies) error {
	var insightsMigrator migrations.TaggedMigrator
	if deps.insightsStore != nil {
		insightsMigrator = insights.NewMigrator(deps.store, deps.insightsStore)
	} else {
		insightsMigrator = insights.NewMigratorNoOp()
	}

	return migrations.RegisterAll(runner, []migrations.TaggedMigrator{
		iam.NewSubscriptionAccountNumberMigrator(deps.store, 500),
		iam.NewLicenseKeyFieldsMigrator(deps.store, 500),
		batches.NewSSHMigratorWithDB(deps.store, deps.keyring.BatchChangesCredentialKey, 5),
		codeintel.NewDiagnosticsCountMigrator(deps.codeIntelStore, 1000),
		codeintel.NewDefinitionLocationsCountMigrator(deps.codeIntelStore, 1000),
		codeintel.NewReferencesLocationsCountMigrator(deps.codeIntelStore, 1000),
		codeintel.NewDocumentColumnSplitMigrator(deps.codeIntelStore, 100),
		codeintel.NewAPIDocsSearchMigrator(),
		insightsMigrator,
	})
}
