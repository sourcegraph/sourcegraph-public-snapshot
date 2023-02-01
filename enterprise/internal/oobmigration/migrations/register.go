package migrations

import (
	"context"
	"database/sql"

	"github.com/derision-test/glock"
	"github.com/sourcegraph/log"

	workerCodeIntel "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared/init/codeintel"
	internalInsights "github.com/sourcegraph/sourcegraph/enterprise/internal/insights"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/batches"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/codeintel"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/iam"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/insights"
	insightsBackfiller "github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/insights/backfillv2"
	insightsrecordingtimes "github.com/sourcegraph/sourcegraph/enterprise/internal/oobmigration/migrations/insights/recording_times"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/oobmigration/migrations"
)

func RegisterEnterpriseMigrators(ctx context.Context, db database.DB, runner *oobmigration.Runner) error {
	codeIntelDB, err := workerCodeIntel.InitRawDB(&observation.TestContext)
	if err != nil {
		return err
	}

	var insightsStore *basestore.Store
	if internalInsights.IsEnabled() {
		codeInsightsDB, err := internalInsights.InitializeCodeInsightsDB(&observation.TestContext, "worker-oobmigrator")
		if err != nil {
			return err
		}

		insightsStore = basestore.NewWithHandle(codeInsightsDB.Handle())
	}

	defaultKeyring := keyring.Default()

	return registerEnterpriseMigrators(runner, false, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: basestore.NewWithHandle(basestore.NewHandleWithDB(log.NoOp(), codeIntelDB, sql.TxOptions{})),
		insightsStore:  insightsStore,
		keyring:        &defaultKeyring,
	})
}

func RegisterEnterpriseMigratorsUsingConfAndStoreFactory(
	ctx context.Context,
	db database.DB,
	runner *oobmigration.Runner,
	conf conftypes.UnifiedQuerier,
	storeFactory migrations.StoreFactory,
) error {
	codeIntelStore, err := storeFactory.Store(ctx, "codeintel")
	if err != nil {
		return err
	}
	insightsStore, err := storeFactory.Store(ctx, "codeinsights")
	if err != nil {
		return err
	}

	keys, err := keyring.NewRing(ctx, conf.SiteConfig().EncryptionKeys)
	if err != nil {
		return err
	}
	if keys == nil {
		keys = &keyring.Ring{}
	}

	return registerEnterpriseMigrators(runner, true, dependencies{
		store:          basestore.NewWithHandle(db.Handle()),
		codeIntelStore: codeIntelStore,
		insightsStore:  insightsStore,
		keyring:        keys,
	})
}

type dependencies struct {
	store          *basestore.Store
	codeIntelStore *basestore.Store
	insightsStore  *basestore.Store
	keyring        *keyring.Ring
}

func registerEnterpriseMigrators(runner *oobmigration.Runner, noDelay bool, deps dependencies) error {
	migrators := []migrations.TaggedMigrator{
		iam.NewSubscriptionAccountNumberMigrator(deps.store, 500),
		iam.NewLicenseKeyFieldsMigrator(deps.store, 500),
		batches.NewSSHMigratorWithDB(deps.store, deps.keyring.BatchChangesCredentialKey, 5),
		codeintel.NewDiagnosticsCountMigrator(deps.codeIntelStore, 1000, 0),
		codeintel.NewDefinitionLocationsCountMigrator(deps.codeIntelStore, 1000, 0),
		codeintel.NewReferencesLocationsCountMigrator(deps.codeIntelStore, 1000, 0),
		codeintel.NewDocumentColumnSplitMigrator(deps.codeIntelStore, 100, 0),
		codeintel.NewSCIPMigrator(deps.store, deps.codeIntelStore),
	}
	if deps.insightsStore != nil {
		migrators = append(migrators,
			insights.NewMigrator(deps.store, deps.insightsStore),
			insightsrecordingtimes.NewRecordingTimesMigrator(deps.insightsStore, 500),
			insightsBackfiller.NewMigrator(deps.insightsStore, glock.NewRealClock(), 10),
		)
	}
	return migrations.RegisterAll(runner, noDelay, migrators)
}
