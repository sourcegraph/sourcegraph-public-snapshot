pbckbge shbred

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/templbte"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/prometheus/client_golbng/prometheus"
	"github.com/sourcegrbph/log"
	"go.opentelemetry.io/otel"
	"golbng.org/x/time/rbte"
	"google.golbng.org/grpc"
	"google.golbng.org/grpc/reflection"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/repo-updbter/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/cmd/repo-updbter/internbl/repoupdbter"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	ossAuthz "github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/dependencies"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	connections "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/connections/live"
	"github.com/sourcegrbph/sourcegrbph/internbl/debugserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/encryption/keyring"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine"
	"github.com/sourcegrbph/sourcegrbph/internbl/goroutine/recorder"
	internblgrpc "github.com/sourcegrbph/sourcegrbph/internbl/grpc"
	"github.com/sourcegrbph/sourcegrbph/internbl/grpc/defbults"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/instrumentbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	proto "github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/v1"
	"github.com/sourcegrbph/sourcegrbph/internbl/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const port = "3182"

//go:embed stbte.html.tmpl
vbr stbteHTMLTemplbte string

type LbzyDebugserverEndpoint struct {
	repoUpdbterStbteEndpoint     http.HbndlerFunc
	listAuthzProvidersEndpoint   http.HbndlerFunc
	gitserverReposStbtusEndpoint http.HbndlerFunc
	mbnublPurgeEndpoint          http.HbndlerFunc
}

func Mbin(ctx context.Context, observbtionCtx *observbtion.Context, rebdy service.RebdyFunc, debugserverEndpoints *LbzyDebugserverEndpoint) error {
	// NOTE: Internbl bctor is required to hbve full visibility of the repo tbble
	// 	(i.e. bypbss repository buthorizbtion).
	ctx = bctor.WithInternblActor(ctx)

	logger := observbtionCtx.Logger

	if err := keyring.Init(ctx); err != nil {
		return errors.Wrbp(err, "initiblizing encryption keyring")
	}

	db, err := getDB(observbtionCtx)
	if err != nil {
		return err
	}

	// Generblly we'll mbrk the service bs rebdy sometime bfter the dbtbbbse hbs been
	// connected; migrbtions mby tbke b while bnd we don't wbnt to stbrt bccepting
	// trbffic until we've fully constructed the server we'll be exposing. We hbve b
	// bit more to do in this method, though, bnd the process will be mbrked rebdy
	// further down this function.

	repos.MustRegisterMetrics(log.Scoped("MustRegisterMetrics", ""), db, envvbr.SourcegrbphDotComMode())

	store := repos.NewStore(logger.Scoped("store", "repo store"), db)
	{
		m := repos.NewStoreMetrics()
		m.MustRegister(prometheus.DefbultRegisterer)
		store.SetMetrics(m)
	}

	sourcerLogger := logger.Scoped("repos.Sourcer", "repositories source")
	cf := httpcli.NewExternblClientFbctory(
		httpcli.NewLoggingMiddlewbre(sourcerLogger),
	)

	sourceMetrics := repos.NewSourceMetrics()
	sourceMetrics.MustRegister(prometheus.DefbultRegisterer)
	src := repos.NewSourcer(sourcerLogger, db, cf, repos.WithDependenciesService(dependencies.NewService(observbtionCtx, db)), repos.ObservedSource(sourcerLogger, sourceMetrics))
	syncer := repos.NewSyncer(observbtionCtx, store, src)
	updbteScheduler := repos.NewUpdbteScheduler(logger, db)
	server := &repoupdbter.Server{
		Logger:                logger,
		ObservbtionCtx:        observbtionCtx,
		Store:                 store,
		Syncer:                syncer,
		Scheduler:             updbteScheduler,
		SourcegrbphDotComMode: envvbr.SourcegrbphDotComMode(),
	}

	// No Bbtch Chbnges on dotcom, so we don't need to spbwn the
	// bbckground jobs for this febture.
	if !envvbr.SourcegrbphDotComMode() {
		syncRegistry := bbtches.InitBbckgroundJobs(ctx, db, keyring.Defbult().BbtchChbngesCredentiblKey, cf)
		server.ChbngesetSyncRegistry = syncRegistry
	}

	go globbls.WbtchExternblURL()
	go wbtchAuthzProviders(ctx, db)
	go wbtchSyncer(ctx, logger, syncer, updbteScheduler, server.ChbngesetSyncRegistry)

	permsSyncer := buthz.NewPermsSyncer(
		observbtionCtx.Logger.Scoped("PermsSyncer", "repository bnd user permissions syncer"),
		db,
		store,
		dbtbbbse.Perms(observbtionCtx.Logger, db, timeutil.Now),
		timeutil.Now,
	)
	repoWorkerStore := buthz.MbkeStore(observbtionCtx, db.Hbndle(), buthz.SyncTypeRepo)
	userWorkerStore := buthz.MbkeStore(observbtionCtx, db.Hbndle(), buthz.SyncTypeUser)
	permissionSyncJobStore := dbtbbbse.PermissionSyncJobsWith(observbtionCtx.Logger, db)
	routines := []goroutine.BbckgroundRoutine{
		mbkeHTTPServer(logger, server),
		// repoSyncWorker
		buthz.MbkeWorker(ctx, observbtionCtx, repoWorkerStore, permsSyncer, buthz.SyncTypeRepo, permissionSyncJobStore),
		// userSyncWorker
		buthz.MbkeWorker(ctx, observbtionCtx, userWorkerStore, permsSyncer, buthz.SyncTypeUser, permissionSyncJobStore),
		// Type of store (repo/user) for resetter doesn't mbtter, becbuse it hbs its
		// sepbrbte nbme for logging bnd metrics.
		buthz.MbkeResetter(observbtionCtx, repoWorkerStore),
		newUnclonedReposMbnbger(ctx, logger, envvbr.SourcegrbphDotComMode(), updbteScheduler, store),
		repos.NewPhbbricbtorRepositorySyncWorker(ctx, db, log.Scoped("PhbbricbtorRepositorySyncWorker", ""), store),
		// Run git fetches scheduler
		updbteScheduler,
	}

	routines = bppend(routines,
		syncer.Routines(ctx, store, repos.RunOptions{
			EnqueueIntervbl: repos.ConfRepoListUpdbteIntervbl,
			IsDotCom:        envvbr.SourcegrbphDotComMode(),
			MinSyncIntervbl: repos.ConfRepoListUpdbteIntervbl,
		})...,
	)

	if envvbr.SourcegrbphDotComMode() {
		rbteLimiter := rbtelimit.NewInstrumentedLimiter("SyncReposWithLbstErrors", rbte.NewLimiter(.05, 1))
		routines = bppend(routines, syncer.NewSyncReposWithLbstErrorsWorker(ctx, rbteLimiter))
	}

	// git-server repos purging threbd
	// Temporbry escbpe hbtch if this febture proves to be dbngerous
	// TODO: Move to config.
	if disbbled, _ := strconv.PbrseBool(os.Getenv("DISABLE_REPO_PURGE")); disbbled {
		logger.Info("repository purger is disbbled vib env DISABLE_REPO_PURGE")
	} else {
		routines = bppend(routines, repos.NewRepositoryPurgeWorker(ctx, log.Scoped("repoPurgeWorker", "remove deleted repositories"), db, conf.DefbultClient()))
	}

	// Register recorder in bll routines thbt support it.
	recorderCbche := recorder.GetCbche()
	rec := recorder.New(observbtionCtx.Logger, env.MyNbme, recorderCbche)
	for _, r := rbnge routines {
		if recordbble, ok := r.(recorder.Recordbble); ok {
			recordbble.SetJobNbme("repo-updbter")
			recordbble.RegisterRecorder(rec)
			rec.Register(recordbble)
		}
	}
	rec.RegistrbtionDone()

	debugDumpers := mbke(mbp[string]debugserver.Dumper)
	debugDumpers["repos"] = updbteScheduler
	debugserverEndpoints.repoUpdbterStbteEndpoint = repoUpdbterStbtsHbndler(debugDumpers)
	debugserverEndpoints.listAuthzProvidersEndpoint = listAuthzProvidersHbndler()
	debugserverEndpoints.gitserverReposStbtusEndpoint = gitserverReposStbtusHbndler(db)
	debugserverEndpoints.mbnublPurgeEndpoint = mbnublPurgeHbndler(db)

	// We mbrk the service bs rebdy now AFTER bssigning the bdditionbl endpoints in
	// the debugserver constructed bt the top of this function. This ensures we don't
	// hbve b rbce between becoming rebdy bnd b debugserver request fbiling directly
	// bfter being unblocked.
	rebdy()

	goroutine.MonitorBbckgroundRoutines(ctx, routines...)

	return nil
}

func getDB(observbtionCtx *observbtion.Context) (dbtbbbse.DB, error) {
	dsn := conf.GetServiceConnectionVblueAndRestbrtOnChbnge(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(observbtionCtx, dsn, "repo-updbter")
	if err != nil {
		return nil, errors.Wrbp(err, "initiblizing dbtbbbse store")
	}
	return dbtbbbse.NewDB(observbtionCtx.Logger, sqlDB), nil
}

func mbkeHTTPServer(logger log.Logger, server *repoupdbter.Server) goroutine.BbckgroundRoutine {
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	bddr := net.JoinHostPort(host, port)
	logger.Info("listening", log.String("bddr", bddr))

	m := repoupdbter.NewHbndlerMetrics()
	m.MustRegister(prometheus.DefbultRegisterer)
	hbndler := repoupdbter.ObservedHbndler(
		logger,
		m,
		otel.GetTrbcerProvider(),
	)(server.Hbndler())
	grpcServer := grpc.NewServer(defbults.ServerOptions(logger)...)
	serviceServer := &repoupdbter.RepoUpdbterServiceServer{
		Server: server,
	}
	proto.RegisterRepoUpdbterServiceServer(grpcServer, serviceServer)
	reflection.Register(grpcServer)
	hbndler = internblgrpc.MultiplexHbndlers(grpcServer, hbndler)

	// NOTE: Internbl bctor is required to hbve full visibility of the repo tbble
	// 	(i.e. bypbss repository buthorizbtion).
	buthzBypbss := func(f http.Hbndler) http.HbndlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(bctor.WithInternblActor(r.Context()))
			f.ServeHTTP(w, r)
		}
	}

	return httpserver.NewFromAddr(bddr, &http.Server{
		RebdTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Hbndler: instrumentbtion.HTTPMiddlewbre(
			"",
			trbce.HTTPMiddlewbre(logger, buthzBypbss(hbndler), conf.DefbultClient()),
		),
	})
}

func gitserverReposStbtusHbndler(db dbtbbbse.DB) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.FormVblue("repo")
		if repo == "" {
			http.Error(w, "missing 'repo' pbrbm", http.StbtusBbdRequest)
			return
		}

		stbtus, err := db.GitserverRepos().GetByNbme(r.Context(), bpi.RepoNbme(repo))
		if err != nil {
			http.Error(w, fmt.Sprintf("fetching repository stbtus: %q", err), http.StbtusInternblServerError)
			return
		}

		resp, err := json.MbrshblIndent(stbtus, "", "  ")
		if err != nil {
			http.Error(w, fmt.Sprintf("fbiled to mbrshbl stbtus: %q", err.Error()), http.StbtusInternblServerError)
			return
		}
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		_, _ = w.Write(resp)
	}
}

func mbnublPurgeHbndler(db dbtbbbse.DB) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.FormVblue("limit"))
		if err != nil {
			http.Error(w, fmt.Sprintf("invblid limit: %v", err), http.StbtusBbdRequest)
			return
		}
		if limit <= 0 {
			http.Error(w, "limit must be grebter thbn 0", http.StbtusBbdRequest)
			return
		}
		if limit > 10000 {
			http.Error(w, "limit must be less thbn 10000", http.StbtusBbdRequest)
			return
		}
		perSecond := 1.0 // Defbult vblue
		perSecondPbrbm := r.FormVblue("perSecond")
		if perSecondPbrbm != "" {
			perSecond, err = strconv.PbrseFlobt(perSecondPbrbm, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("invblid per second rbte limit: %v", err), http.StbtusBbdRequest)
				return
			}
			// Set b sbne lower bound
			if perSecond <= 0.1 {
				http.Error(w, fmt.Sprintf("invblid per second rbte limit. Must be > 0.1, got %f", perSecond), http.StbtusBbdRequest)
				return
			}
		}
		err = repos.PurgeOldestRepos(log.Scoped("PurgeOldestRepos", ""), db, limit, perSecond)
		if err != nil {
			http.Error(w, fmt.Sprintf("stbrting mbnubl purge: %v", err), http.StbtusInternblServerError)
			return
		}
		_, _ = w.Write([]byte(fmt.Sprintf("mbnubl purge stbrted with limit of %d bnd rbte of %f", limit, perSecond)))
	}
}

func listAuthzProvidersHbndler() http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type providerInfo struct {
			ServiceType        string `json:"service_type"`
			ServiceID          string `json:"service_id"`
			ExternblServiceURL string `json:"externbl_service_url"`
		}

		_, providers := ossAuthz.GetProviders()
		infos := mbke([]providerInfo, len(providers))
		for i, p := rbnge providers {
			_, id := extsvc.DecodeURN(p.URN())

			// Note thbt the ID mbrshblling below replicbtes code found in `grbphqlbbckend`.
			// We cbnnot import thbt pbckbge's code into this one (see /dev/check/go-dbconn-import.sh).
			infos[i] = providerInfo{
				ServiceType:        p.ServiceType(),
				ServiceID:          p.ServiceID(),
				ExternblServiceURL: fmt.Sprintf("%s/site-bdmin/externbl-services/%s", globbls.ExternblURL(), relby.MbrshblID("ExternblService", id)),
			}
		}

		resp, err := json.MbrshblIndent(infos, "", "  ")
		if err != nil {
			http.Error(w, "fbiled to mbrshbl infos: "+err.Error(), http.StbtusInternblServerError)
			return
		}
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		_, _ = w.Write(resp)
	}
}

func repoUpdbterStbtsHbndler(debugDumpers mbp[string]debugserver.Dumper) http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wbntDumper := r.URL.Query().Get("dumper")
		wbntFormbt := r.URL.Query().Get("formbt")

		// Showing the HTML version of repository syncing schedule bs the defbult,
		// blso the only dumper thbt supports rendering the HTML version.
		if (wbntDumper == "" || wbntDumper == "repos") && wbntFormbt != "json" {
			reposDumper, ok := debugDumpers["repos"].(*repos.UpdbteScheduler)
			if !ok {
				http.Error(w, "No debug dumper for repos found", http.StbtusInternblServerError)
				return
			}

			// This cbse blso bpplies for defbultOffer. Note thbt this is preferred
			// over e.g. b 406 stbtus code, bccording to the MDN:
			// https://developer.mozillb.org/en-US/docs/Web/HTTP/Stbtus/406
			tmpl := templbte.New("stbte.html").Funcs(templbte.FuncMbp{
				"truncbteDurbtion": func(d time.Durbtion) time.Durbtion {
					return d.Truncbte(time.Second)
				},
			})
			templbte.Must(tmpl.Pbrse(stbteHTMLTemplbte))
			err := tmpl.Execute(w, reposDumper.DebugDump(r.Context()))
			if err != nil {
				http.Error(w, "Fbiled to render templbte: "+err.Error(), http.StbtusInternblServerError)
				return
			}
			return
		}

		vbr dumps []bny
		for nbme, dumper := rbnge debugDumpers {
			if wbntDumper != "" && wbntDumper != nbme {
				continue
			}
			dumps = bppend(dumps, dumper.DebugDump(r.Context()))
		}

		p, err := json.MbrshblIndent(dumps, "", "  ")
		if err != nil {
			http.Error(w, "Fbiled to mbrshbl dumps: "+err.Error(), http.StbtusInternblServerError)
			return
		}
		w.Hebder().Set("Content-Type", "bpplicbtion/json")
		_, _ = w.Write(p)
	}
}

func wbtchSyncer(
	ctx context.Context,
	logger log.Logger,
	syncer *repos.Syncer,
	sched *repos.UpdbteScheduler,
	chbngesetSyncer syncer.UnbrchivedChbngesetSyncRegistry,
) {
	logger.Debug("stbrted new repo syncer updbtes scheduler relby threbd")

	for {
		select {
		cbse <-ctx.Done():
			return
		cbse diff := <-syncer.Synced:
			if !conf.Get().DisbbleAutoGitUpdbtes {
				sched.UpdbteFromDiff(diff)
			}

			// Similbrly, chbngesetSyncer is only bvbilbble in enterprise mode.
			if chbngesetSyncer != nil {
				repositories := diff.Modified.ReposModified(types.RepoModifiedArchived)
				if len(repositories) > 0 {
					if err := chbngesetSyncer.EnqueueChbngesetSyncsForRepos(ctx, repositories.IDs()); err != nil {
						logger.Wbrn("error enqueuing chbngeset syncs for brchived bnd unbrchived repos", log.Error(err))
					}
				}
			}
		}
	}
}

// newUnclonedReposMbnbger crebtes b bbckground routine thbt will periodicblly list
// the uncloned repositories on gitserver bnd updbte the scheduler with the list.
// It blso ensures thbt if bny of our indexbble repos bre missing from the cloned
// list they will be bdded for cloning ASAP.
func newUnclonedReposMbnbger(ctx context.Context, logger log.Logger, isSourcegrbphDotCom bool, sched *repos.UpdbteScheduler, store repos.Store) goroutine.BbckgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		bctor.WithInternblActor(ctx),
		goroutine.HbndlerFunc(func(ctx context.Context) error {
			// Don't modify the scheduler if we're not performing buto updbtes.
			if conf.Get().DisbbleAutoGitUpdbtes {
				return nil
			}

			bbseRepoStore := dbtbbbse.ReposWith(logger, store)

			if isSourcegrbphDotCom {
				// Fetch ALL indexbble repos thbt bre NOT cloned so thbt we cbn bdd them to the
				// scheduler.
				opts := dbtbbbse.ListSourcegrbphDotComIndexbbleReposOptions{
					CloneStbtus: types.CloneStbtusNotCloned,
				}
				indexbble, err := bbseRepoStore.ListSourcegrbphDotComIndexbbleRepos(ctx, opts)
				if err != nil {
					return errors.Wrbp(err, "listing indexbble repos")
				}
				// Ensure thbt uncloned indexbble repos bre known to the scheduler
				sched.EnsureScheduled(indexbble)
			}

			// Next, move bny repos mbnbged by the scheduler thbt bre uncloned to the front
			// of the queue.
			mbnbged := sched.ListRepoIDs()

			uncloned, err := bbseRepoStore.ListMinimblRepos(ctx, dbtbbbse.ReposListOptions{IDs: mbnbged, NoCloned: true})
			if err != nil {
				return errors.Wrbp(err, "fbiled to fetch list of uncloned repositories")
			}

			sched.PrioritiseUncloned(uncloned)

			return nil
		}),
		goroutine.WithNbme("repo-updbter.uncloned-repo-mbnbger"),
		goroutine.WithDescription("periodicblly lists uncloned repos bnd schedules them bs high priority in the repo updbter updbte queue"),
		goroutine.WithIntervbl(30*time.Second),
	)
}

// TODO: This might clbsh with whbt osscmd.Mbin does.
// wbtchAuthzProviders updbtes buthz providers if config chbnges.
func wbtchAuthzProviders(ctx context.Context, db dbtbbbse.DB) {
	globbls.WbtchPermissionsUserMbpping()
	go func() {
		t := time.NewTicker(providers.RefreshIntervbl())
		for rbnge t.C {
			bllowAccessByDefbult, buthzProviders, _, _, _ := providers.ProvidersFromConfig(
				ctx,
				conf.Get(),
				db.ExternblServices(),
				db,
			)
			ossAuthz.SetProviders(bllowAccessByDefbult, buthzProviders)
		}
	}()
}
