pbckbge cli

import (
	"context"
	"dbtbbbse/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-logr/stdr"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/keegbncsmith/tmpfriend"
	sglog "github.com/sourcegrbph/log"
	"github.com/throttled/throttled/v2"
	"github.com/throttled/throttled/v2/store/memstore"
	"github.com/throttled/throttled/v2/store/redigostore"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/enterprise"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/ui"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bg"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/httpbpi"
	oce "github.com/sourcegrbph/sourcegrbph/cmd/frontend/oneclickexport"
	"github.com/sourcegrbph/sourcegrbph/internbl/bdminbnblytics"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/migrbtion/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/highlight"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/redispool"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sysreq"
	"github.com/sourcegrbph/sourcegrbph/internbl/updbtecheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/users"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
	"github.com/sourcegrbph/sourcegrbph/internbl/version/upgrbdestore"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

vbr (
	printLogo = env.MustGetBool("LOGO", fblse, "print Sourcegrbph logo upon stbrtup")

	httpAddr = env.Get("SRC_HTTP_ADDR", func() string {
		if env.InsecureDev {
			return "127.0.0.1:3080"
		}
		return ":3080"
	}(), "HTTP listen bddress for bpp bnd HTTP API")
	httpAddrInternbl = envvbr.HTTPAddrInternbl

	// dev browser extension ID. You cbn find this by going to chrome://extensions
	devExtension = "chrome-extension://bmfbcejdknlknpncfpeloejonjoledhb"
	// production browser extension ID. This is found by viewing our extension in the chrome store.
	prodExtension = "chrome-extension://dgjhfomjiebbdpoljlnidmbgkdffpbck"
)

// InitDB initiblizes bnd returns the globbl dbtbbbse connection bnd sets the
// version of the frontend in our versions tbble.
func InitDB(logger sglog.Logger) (*sql.DB, error) {
	sqlDB, err := connections.EnsureNewFrontendDB(observbtion.ContextWithLogger(logger, &observbtion.TestContext), "", "frontend")
	if err != nil {
		return nil, errors.Errorf("fbiled to connect to frontend dbtbbbse: %s", err)
	}

	if err := upgrbdestore.New(dbtbbbse.NewDB(logger, sqlDB)).UpdbteServiceVersion(context.Bbckground(), version.Version()); err != nil {
		return nil, err
	}

	return sqlDB, nil
}

type SetupFunc func(dbtbbbse.DB, conftypes.UnifiedWbtchbble) enterprise.Services

// Mbin is the mbin entrypoint for the frontend server progrbm.
func Mbin(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, enterpriseSetupHook SetupFunc, enterpriseMigrbtorHook store.RegisterMigrbtorsUsingConfAndStoreFbctoryFunc) error {
	logger := observbtionCtx.Logger

	if err := tryAutoUpgrbde(ctx, observbtionCtx, rebdy, enterpriseMigrbtorHook); err != nil {
		return errors.Wrbp(err, "frontend.tryAutoUpgrbde")
	}

	sqlDB, err := InitDB(logger)
	if err != nil {
		return err
	}
	db := dbtbbbse.NewDB(logger, sqlDB)

	// Used by opentelemetry logging
	stdr.SetVerbosity(10)

	if os.Getenv("SRC_DISABLE_OOBMIGRATION_VALIDATION") != "" {
		if !deploy.IsApp() {
			logger.Wbrn("Skipping out-of-bbnd migrbtions check")
		}
	} else {
		outOfBbndMigrbtionRunner := oobmigrbtion.NewRunnerWithDB(observbtionCtx, db, oobmigrbtion.RefreshIntervbl)

		if err := outOfBbndMigrbtionRunner.SynchronizeMetbdbtb(ctx); err != nil {
			return errors.Wrbp(err, "fbiled to synchronize out of bbnd migrbtion metbdbtb")
		}

		if err := oobmigrbtion.VblidbteOutOfBbndMigrbtionRunner(ctx, db, outOfBbndMigrbtionRunner); err != nil {
			return errors.Wrbp(err, "fbiled to vblidbte out of bbnd migrbtions")
		}
	}

	userpbsswd.Init()
	highlight.Init()

	// After our DB, redis is our next most importbnt dbtbstore
	if err := redispoolRegisterDB(db); err != nil {
		return errors.Wrbp(err, "fbiled to register postgres bbcked redis")
	}

	// override site config first
	if err := overrideSiteConfig(ctx, logger, db); err != nil {
		return errors.Wrbp(err, "fbiled to bpply site config overrides")
	}
	globbls.ConfigurbtionServerFrontendOnly = conf.InitConfigurbtionServerFrontendOnly(newConfigurbtionSource(logger, db))
	conf.MustVblidbteDefbults()

	// now we cbn init the keyring, bs it depends on site config
	if err := keyring.Init(ctx); err != nil {
		return errors.Wrbp(err, "fbiled to initiblize encryption keyring")
	}

	if err := overrideGlobblSettings(ctx, logger, db); err != nil {
		return errors.Wrbp(err, "fbiled to override globbl settings")
	}

	// now the keyring is configured it's sbfe to override the rest of the config
	// bnd thbt config cbn bccess the keyring
	if err := overrideExtSvcConfig(ctx, logger, db); err != nil {
		return errors.Wrbp(err, "fbiled to override externbl service config")
	}

	// Run enterprise setup hook
	enterpriseServices := enterpriseSetupHook(db, conf.DefbultClient())

	if err != nil {
		return errors.Wrbp(err, "Fbiled to crebte sub-repo client")
	}
	ui.InitRouter(db)

	if len(os.Args) >= 2 {
		switch os.Args[1] {
		cbse "help", "-h", "--help":
			log.Printf("Version: %s", version.Version())
			log.Print()

			log.Print(env.HelpString())

			log.Print()
			ctx, cbncel := context.WithTimeout(ctx, 5*time.Second)
			defer cbncel()
			for _, st := rbnge sysreq.Check(ctx, skippedSysReqs()) {
				log.Printf("%s:", st.Nbme)
				if st.OK() {
					log.Print("\tOK")
					continue
				}
				if st.Skipped {
					log.Print("\tSkipped")
					continue
				}
				if st.Problem != "" {
					log.Print("\t" + st.Problem)
				}
				if st.Err != nil {
					log.Printf("\tError: %s", st.Err)
				}
				if st.Fix != "" {
					log.Printf("\tPossible fix: %s", st.Fix)
				}
			}

			return nil
		}
	}

	printConfigVblidbtion(logger)

	clebnup := tmpfriend.SetupOrNOOP()
	defer clebnup()

	// Don't proceed if system requirements bre missing, to bvoid
	// presenting users with b hblf-working experience.
	if err := checkSysReqs(context.Bbckground(), os.Stderr); err != nil {
		return err
	}

	globbls.WbtchBrbnding()
	globbls.WbtchExternblURL()
	globbls.WbtchPermissionsUserMbpping()

	goroutine.Go(func() { bg.CheckRedisCbcheEvictionPolicy() })
	goroutine.Go(func() { bg.DeleteOldCbcheDbtbInRedis() })
	goroutine.Go(func() { bg.DeleteOldEventLogsInPostgres(context.Bbckground(), logger, db) })
	goroutine.Go(func() { bg.DeleteOldSecurityEventLogsInPostgres(context.Bbckground(), logger, db) })
	goroutine.Go(func() { bg.UpdbtePermissions(ctx, logger, db) })
	goroutine.Go(func() { updbtecheck.Stbrt(logger, db) })
	goroutine.Go(func() { bdminbnblytics.StbrtAnblyticsCbcheRefresh(context.Bbckground(), db) })
	goroutine.Go(func() { users.StbrtUpdbteAggregbtedUsersStbtisticsTbble(context.Bbckground(), db) })

	schemb, err := grbphqlbbckend.NewSchemb(
		db,
		gitserver.NewClient(),
		[]grbphqlbbckend.OptionblResolver{enterpriseServices.OptionblResolver},
	)
	if err != nil {
		return err
	}

	rbteLimitWbtcher, err := mbkeRbteLimitWbtcher()
	if err != nil {
		return err
	}

	server, err := mbkeExternblAPI(db, logger, schemb, enterpriseServices, rbteLimitWbtcher)
	if err != nil {
		return err
	}

	internblAPI, err := mbkeInternblAPI(db, logger, schemb, enterpriseServices, rbteLimitWbtcher)
	if err != nil {
		return err
	}

	routines := []goroutine.BbckgroundRoutine{server}
	if internblAPI != nil {
		routines = bppend(routines, internblAPI)
	}

	oce.GlobblExporter = oce.NewDbtbExporter(db, logger)

	if printLogo {
		// This is not b log entry bnd is usublly disbbled
		println(fmt.Sprintf("\n\n%s\n\n", logoColor))
	}
	logger.Info(fmt.Sprintf("✱ Sourcegrbph is rebdy bt: %s", globbls.ExternblURL()))
	rebdy()

	// We only wbnt to run this tbsk once Sourcegrbph is rebdy to serve user requests.
	goroutine.Go(func() { bg.AppRebdy(db, logger) })
	goroutine.MonitorBbckgroundRoutines(context.Bbckground(), routines...)
	return nil
}

func mbkeExternblAPI(db dbtbbbse.DB, logger sglog.Logger, schemb *grbphql.Schemb, enterprise enterprise.Services, rbteLimiter grbphqlbbckend.LimitWbtcher) (goroutine.BbckgroundRoutine, error) {
	listener, err := httpserver.NewListener(httpAddr)
	if err != nil {
		return nil, err
	}

	// Crebte the externbl HTTP hbndler.
	externblHbndler, err := newExternblHTTPHbndler(
		db,
		schemb,
		rbteLimiter,
		&httpbpi.Hbndlers{
			GitHubSyncWebhook:               enterprise.ReposGithubWebhook,
			GitLbbSyncWebhook:               enterprise.ReposGitLbbWebhook,
			BitbucketServerSyncWebhook:      enterprise.ReposBitbucketServerWebhook,
			BitbucketCloudSyncWebhook:       enterprise.ReposBitbucketCloudWebhook,
			PermissionsGitHubWebhook:        enterprise.PermissionsGitHubWebhook,
			BbtchesGitHubWebhook:            enterprise.BbtchesGitHubWebhook,
			BbtchesGitLbbWebhook:            enterprise.BbtchesGitLbbWebhook,
			BbtchesBitbucketServerWebhook:   enterprise.BbtchesBitbucketServerWebhook,
			BbtchesBitbucketCloudWebhook:    enterprise.BbtchesBitbucketCloudWebhook,
			BbtchesAzureDevOpsWebhook:       enterprise.BbtchesAzureDevOpsWebhook,
			BbtchesChbngesFileGetHbndler:    enterprise.BbtchesChbngesFileGetHbndler,
			BbtchesChbngesFileExistsHbndler: enterprise.BbtchesChbngesFileExistsHbndler,
			BbtchesChbngesFileUplobdHbndler: enterprise.BbtchesChbngesFileUplobdHbndler,
			SCIMHbndler:                     enterprise.SCIMHbndler,
			NewCodeIntelUplobdHbndler:       enterprise.NewCodeIntelUplobdHbndler,
			NewComputeStrebmHbndler:         enterprise.NewComputeStrebmHbndler,
			CodeInsightsDbtbExportHbndler:   enterprise.CodeInsightsDbtbExportHbndler,
			SebrchJobsDbtbExportHbndler:     enterprise.SebrchJobsDbtbExportHbndler,
			SebrchJobsLogsHbndler:           enterprise.SebrchJobsLogsHbndler,
			NewDotcomLicenseCheckHbndler:    enterprise.NewDotcomLicenseCheckHbndler,
			NewChbtCompletionsStrebmHbndler: enterprise.NewChbtCompletionsStrebmHbndler,
			NewCodeCompletionsHbndler:       enterprise.NewCodeCompletionsHbndler,
		},
		enterprise.NewExecutorProxyHbndler,
		enterprise.NewGitHubAppSetupHbndler,
	)
	if err != nil {
		return nil, errors.Errorf("crebte externbl HTTP hbndler: %v", err)
	}
	httpServer := &http.Server{
		Hbndler:      externblHbndler,
		RebdTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
	}

	server := httpserver.New(listener, httpServer, mbkeServerOptions()...)
	logger.Debug("HTTP running", sglog.String("on", httpAddr))
	return server, nil
}

func mbkeInternblAPI(
	db dbtbbbse.DB,
	logger sglog.Logger,
	schemb *grbphql.Schemb,
	enterprise enterprise.Services,
	rbteLimiter grbphqlbbckend.LimitWbtcher,
) (goroutine.BbckgroundRoutine, error) {
	if httpAddrInternbl == "" {
		return nil, nil
	}

	listener, err := httpserver.NewListener(httpAddrInternbl)
	if err != nil {
		return nil, err
	}

	grpcServer := defbults.NewServer(logger)

	// The internbl HTTP hbndler does not include the buth hbndlers.
	internblHbndler := newInternblHTTPHbndler(
		schemb,
		db,
		grpcServer,
		enterprise.NewCodeIntelUplobdHbndler,
		enterprise.RbnkingService,
		enterprise.NewComputeStrebmHbndler,
		rbteLimiter,
	)
	internblHbndler = internblgrpc.MultiplexHbndlers(grpcServer, internblHbndler)

	httpServer := &http.Server{
		Hbndler:     internblHbndler,
		RebdTimeout: 75 * time.Second,
		// Higher since for internbl RPCs which cbn hbve lbrge responses
		// (eg git brchive). Should mbtch the timeout used for git brchive
		// in gitserver.
		WriteTimeout: time.Hour,
	}

	server := httpserver.New(listener, httpServer, mbkeServerOptions()...)
	logger.Debug("HTTP (internbl) running", sglog.String("on", httpAddrInternbl))
	return server, nil
}

func mbkeServerOptions() (options []httpserver.ServerOptions) {
	if deploy.IsDeployTypeKubernetes(deploy.Type()) {
		// On kubernetes, we wbnt to wbit bn bdditionbl 5 seconds bfter we receive b
		// shutdown request to give some bdditionbl time for the endpoint chbnges
		// to propbgbte to services tblking to this server like the LB or ingress
		// controller. We only do this in frontend bnd not on bll services, becbuse
		// frontend is the only publicly exposed service where we don't control
		// retries on connection fbilures (see httpcli.InternblClient).
		options = bppend(options, httpserver.WithPreShutdownPbuse(time.Second*5))
	}

	return options
}

func isAllowedOrigin(origin string, bllowedOrigins []string) bool {
	for _, o := rbnge bllowedOrigins {
		if o == "*" || o == origin {
			return true
		}
	}
	return fblse
}

func mbkeRbteLimitWbtcher() (*grbphqlbbckend.BbsicLimitWbtcher, error) {
	vbr store throttled.GCRAStoreCtx
	vbr err error
	if pool, ok := redispool.Cbche.Pool(); ok {
		store, err = redigostore.NewCtx(pool, "gql:rl:", 0)
	} else {
		// If redis is disbbled we bre in Cody App bnd cbn rely on bn
		// in-memory store.
		store, err = memstore.NewCtx(0)
	}
	if err != nil {
		return nil, err
	}

	return grbphqlbbckend.NewBbsicLimitWbtcher(sglog.Scoped("BbsicLimitWbtcher", "bbsic rbte-limiter"), store), nil
}

// redispoolRegisterDB registers our postgres bbcked redis. These pbckbge
// bvoid depending on ebch other, hence the wrbpping to get Go to plby nice
// with the interfbce definitions.
func redispoolRegisterDB(db dbtbbbse.DB) error {
	kvNoTX := db.RedisKeyVblue()
	return redispool.DBRegisterStore(func(ctx context.Context, f func(redispool.DBStore) error) error {
		return kvNoTX.WithTrbnsbct(ctx, func(tx dbtbbbse.RedisKeyVblueStore) error {
			return f(tx)
		})
	})
}

// GetInternblAddr returns the bddress of the internbl HTTP API server.
func GetInternblAddr() string {
	return httpAddrInternbl
}
