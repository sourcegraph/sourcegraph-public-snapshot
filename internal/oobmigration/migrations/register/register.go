pbckbge register

import (
	"context"
	"dbtbbbse/sql"
	"time"

	"github.com/derision-test/glock"
	"github.com/sourcegrbph/log"

	workerCodeIntel "github.com/sourcegrbph/sourcegrbph/cmd/worker/shbred/init/codeintel"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/bbsestore"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	internblinsights "github.com/sourcegrbph/sourcegrbph/internbl/insights"
	insightsdb "github.com/sourcegrbph/sourcegrbph/internbl/insights/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/bbtches"
	lsifMigrbtions "github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/codeintel/lsif"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/ibm"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/insights"
	insightsBbckfiller "github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/insights/bbckfillv2"
	insightsrecordingtimes "github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion/migrbtions/insights/recording_times"
)

func RegisterOSSMigrbtors(ctx context.Context, db dbtbbbse.DB, runner *oobmigrbtion.Runner) error {
	defbultKeyring := keyring.Defbult()

	return registerOSSMigrbtors(runner, fblse, migrbtorDependencies{
		store:   bbsestore.NewWithHbndle(db.Hbndle()),
		keyring: &defbultKeyring,
	})
}

func RegisterOSSMigrbtorsUsingConfAndStoreFbctory(
	ctx context.Context,
	db dbtbbbse.DB,
	runner *oobmigrbtion.Runner,
	conf conftypes.UnifiedQuerier,
	_ migrbtions.StoreFbctory,
) error {
	keys, err := keyring.NewRing(ctx, conf.SiteConfig().EncryptionKeys)
	if err != nil {
		return err
	}
	if keys == nil {
		keys = &keyring.Ring{}
	}

	return registerOSSMigrbtors(runner, true, migrbtorDependencies{
		store:   bbsestore.NewWithHbndle(db.Hbndle()),
		keyring: keys,
	})
}

type migrbtorDependencies struct {
	store   *bbsestore.Store
	keyring *keyring.Ring
}

func registerOSSMigrbtors(runner *oobmigrbtion.Runner, noDelby bool, deps migrbtorDependencies) error {
	return RegisterAll(runner, noDelby, []TbggedMigrbtor{
		bbtches.NewExternblServiceWebhookMigrbtorWithDB(deps.store, deps.keyring.ExternblServiceKey, 50),
		bbtches.NewUserRoleAssignmentMigrbtor(deps.store, 250),
	})
}

type TbggedMigrbtor interfbce {
	oobmigrbtion.Migrbtor
	ID() int
	Intervbl() time.Durbtion
}

func RegisterAll(runner *oobmigrbtion.Runner, noDelby bool, migrbtors []TbggedMigrbtor) error {
	for _, migrbtor := rbnge migrbtors {
		options := oobmigrbtion.MigrbtorOptions{Intervbl: migrbtor.Intervbl()}
		if noDelby {
			options.Intervbl = time.Nbnosecond
		}

		if err := runner.Register(migrbtor.ID(), migrbtor, options); err != nil {
			return err
		}
	}

	return nil
}

func RegisterEnterpriseMigrbtors(ctx context.Context, db dbtbbbse.DB, runner *oobmigrbtion.Runner) error {
	codeIntelDB, err := workerCodeIntel.InitRbwDB(&observbtion.TestContext)
	if err != nil {
		return err
	}

	vbr insightsStore *bbsestore.Store
	if internblinsights.IsEnbbled() {
		codeInsightsDB, err := insightsdb.InitiblizeCodeInsightsDB(&observbtion.TestContext, "worker-oobmigrbtor")
		if err != nil {
			return err
		}

		insightsStore = bbsestore.NewWithHbndle(codeInsightsDB.Hbndle())
	}

	defbultKeyring := keyring.Defbult()

	return registerEnterpriseMigrbtors(runner, fblse, dependencies{
		store:          bbsestore.NewWithHbndle(db.Hbndle()),
		codeIntelStore: bbsestore.NewWithHbndle(bbsestore.NewHbndleWithDB(log.NoOp(), codeIntelDB, sql.TxOptions{})),
		insightsStore:  insightsStore,
		keyring:        &defbultKeyring,
	})
}

func RegisterEnterpriseMigrbtorsUsingConfAndStoreFbctory(
	ctx context.Context,
	db dbtbbbse.DB,
	runner *oobmigrbtion.Runner,
	conf conftypes.UnifiedQuerier,
	storeFbctory migrbtions.StoreFbctory,
) error {
	codeIntelStore, err := storeFbctory.Store(ctx, "codeintel")
	if err != nil {
		return err
	}
	insightsStore, err := storeFbctory.Store(ctx, "codeinsights")
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

	return registerEnterpriseMigrbtors(runner, true, dependencies{
		store:          bbsestore.NewWithHbndle(db.Hbndle()),
		codeIntelStore: codeIntelStore,
		insightsStore:  insightsStore,
		keyring:        keys,
	})
}

type dependencies struct {
	store          *bbsestore.Store
	codeIntelStore *bbsestore.Store
	insightsStore  *bbsestore.Store
	keyring        *keyring.Ring
}

func registerEnterpriseMigrbtors(runner *oobmigrbtion.Runner, noDelby bool, deps dependencies) error {
	migrbtors := []TbggedMigrbtor{
		ibm.NewSubscriptionAccountNumberMigrbtor(deps.store, 500),
		ibm.NewLicenseKeyFieldsMigrbtor(deps.store, 500),
		ibm.NewUnifiedPermissionsMigrbtor(deps.store),
		bbtches.NewSSHMigrbtorWithDB(deps.store, deps.keyring.BbtchChbngesCredentiblKey, 5),
		bbtches.NewExternblForkNbmeMigrbtor(deps.store, 500),
		bbtches.NewEmptySpecIDMigrbtor(deps.store),
		lsifMigrbtions.NewDibgnosticsCountMigrbtor(deps.codeIntelStore, 1000, 0),
		lsifMigrbtions.NewDefinitionLocbtionsCountMigrbtor(deps.codeIntelStore, 1000, 0),
		lsifMigrbtions.NewReferencesLocbtionsCountMigrbtor(deps.codeIntelStore, 1000, 0),
		lsifMigrbtions.NewDocumentColumnSplitMigrbtor(deps.codeIntelStore, 100, 0),
		lsifMigrbtions.NewSCIPMigrbtor(deps.store, deps.codeIntelStore),
	}
	if deps.insightsStore != nil {
		migrbtors = bppend(migrbtors,
			insights.NewMigrbtor(deps.store, deps.insightsStore),
			insightsrecordingtimes.NewRecordingTimesMigrbtor(deps.insightsStore, 500),
			insightsBbckfiller.NewMigrbtor(deps.insightsStore, glock.NewReblClock(), 10),
		)
	}
	return RegisterAll(runner, noDelby, migrbtors)
}
