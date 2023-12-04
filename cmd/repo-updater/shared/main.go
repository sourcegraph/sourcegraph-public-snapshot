package shared

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"html/template"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/graph-gophers/graphql-go/relay"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/authz"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/repoupdater"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	ossAuthz "github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/authz/providers"
	"github.com/sourcegraph/sourcegraph/internal/batches"
	"github.com/sourcegraph/sourcegraph/internal/batches/syncer"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/dependencies"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/goroutine/recorder"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/repos"
	"github.com/sourcegraph/sourcegraph/internal/repos/scheduler"
	proto "github.com/sourcegraph/sourcegraph/internal/repoupdater/v1"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/timeutil"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const port = "3182"

//go:embed state.html.tmpl
var stateHTMLTemplate string

type LazyDebugserverEndpoint struct {
	repoUpdaterStateEndpoint     http.HandlerFunc
	listAuthzProvidersEndpoint   http.HandlerFunc
	gitserverReposStatusEndpoint http.HandlerFunc
	manualPurgeEndpoint          http.HandlerFunc
}

func Main(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, debugserverEndpoints *LazyDebugserverEndpoint) error {
	// NOTE: Internal actor is required to have full visibility of the repo table
	// 	(i.e. bypass repository authorization).
	ctx = actor.WithInternalActor(ctx)

	logger := observationCtx.Logger

	if err := keyring.Init(ctx); err != nil {
		return errors.Wrap(err, "initializing encryption keyring")
	}

	db, err := getDB(observationCtx)
	if err != nil {
		return err
	}

	// Generally we'll mark the service as ready sometime after the database has been
	// connected; migrations may take a while and we don't want to start accepting
	// traffic until we've fully constructed the server we'll be exposing. We have a
	// bit more to do in this method, though, and the process will be marked ready
	// further down this function.

	repos.MustRegisterMetrics(log.Scoped("MustRegisterMetrics"), db, envvar.SourcegraphDotComMode())

	store := repos.NewStore(logger.Scoped("store"), db)
	{
		m := repos.NewStoreMetrics()
		m.MustRegister(prometheus.DefaultRegisterer)
		store.SetMetrics(m)
	}

	sourcerLogger := logger.Scoped("repos.Sourcer")
	cf := httpcli.NewExternalClientFactory(
		httpcli.NewLoggingMiddleware(sourcerLogger),
	)

	sourceMetrics := repos.NewSourceMetrics()
	sourceMetrics.MustRegister(prometheus.DefaultRegisterer)
	src := repos.NewSourcer(sourcerLogger, db, cf, repos.WithDependenciesService(dependencies.NewService(observationCtx, db)), repos.ObservedSource(sourcerLogger, sourceMetrics))
	syncer := repos.NewSyncer(observationCtx, store, src)
	updateScheduler := scheduler.NewUpdateScheduler(logger, db, gitserver.NewClient("repos.updatescheduler"))
	server := &repoupdater.Server{
		Logger:                logger,
		ObservationCtx:        observationCtx,
		Store:                 store,
		Syncer:                syncer,
		Scheduler:             updateScheduler,
		SourcegraphDotComMode: envvar.SourcegraphDotComMode(),
	}

	// No Batch Changes on dotcom, so we don't need to spawn the
	// background jobs for this feature.
	if !envvar.SourcegraphDotComMode() {
		syncRegistry := batches.InitBackgroundJobs(ctx, db, keyring.Default().BatchChangesCredentialKey, cf)
		server.ChangesetSyncRegistry = syncRegistry
	}

	go globals.WatchExternalURL()
	go watchAuthzProviders(ctx, db)
	go watchSyncer(ctx, logger, syncer, updateScheduler, server.ChangesetSyncRegistry)

	permsSyncer := authz.NewPermsSyncer(
		observationCtx.Logger.Scoped("PermsSyncer"),
		db,
		store,
		database.Perms(observationCtx.Logger, db, timeutil.Now),
		timeutil.Now,
	)
	repoWorkerStore := authz.MakeStore(observationCtx, db.Handle(), authz.SyncTypeRepo)
	userWorkerStore := authz.MakeStore(observationCtx, db.Handle(), authz.SyncTypeUser)
	permissionSyncJobStore := database.PermissionSyncJobsWith(observationCtx.Logger, db)
	routines := []goroutine.BackgroundRoutine{
		makeHTTPServer(logger, server),
		// repoSyncWorker
		authz.MakeWorker(ctx, observationCtx, repoWorkerStore, permsSyncer, authz.SyncTypeRepo, permissionSyncJobStore),
		// userSyncWorker
		authz.MakeWorker(ctx, observationCtx, userWorkerStore, permsSyncer, authz.SyncTypeUser, permissionSyncJobStore),
		// Type of store (repo/user) for resetter doesn't matter, because it has its
		// separate name for logging and metrics.
		authz.MakeResetter(observationCtx, repoWorkerStore),
		newUnclonedReposManager(ctx, logger, envvar.SourcegraphDotComMode(), updateScheduler, store),
		repos.NewPhabricatorRepositorySyncWorker(ctx, db, log.Scoped("PhabricatorRepositorySyncWorker"), store),
		// Run git fetches scheduler
		updateScheduler,
	}

	routines = append(routines,
		syncer.Routines(ctx, store, repos.RunOptions{
			EnqueueInterval: repos.ConfRepoListUpdateInterval,
			IsDotCom:        envvar.SourcegraphDotComMode(),
			MinSyncInterval: repos.ConfRepoListUpdateInterval,
		})...,
	)

	if envvar.SourcegraphDotComMode() {
		rateLimiter := ratelimit.NewInstrumentedLimiter("SyncReposWithLastErrors", rate.NewLimiter(1, 1))
		routines = append(routines, syncer.NewSyncReposWithLastErrorsWorker(ctx, rateLimiter))
	}

	// git-server repos purging thread
	// Temporary escape hatch if this feature proves to be dangerous
	// TODO: Move to config.
	if disabled, _ := strconv.ParseBool(os.Getenv("DISABLE_REPO_PURGE")); disabled {
		logger.Info("repository purger is disabled via env DISABLE_REPO_PURGE")
	} else {
		routines = append(routines, repos.NewRepositoryPurgeWorker(ctx, log.Scoped("repoPurgeWorker"), db, conf.DefaultClient()))
	}

	// Register recorder in all routines that support it.
	recorderCache := recorder.GetCache()
	rec := recorder.New(observationCtx.Logger, env.MyName, recorderCache)
	for _, r := range routines {
		if recordable, ok := r.(recorder.Recordable); ok {
			recordable.SetJobName("repo-updater")
			recordable.RegisterRecorder(rec)
			rec.Register(recordable)
		}
	}
	rec.RegistrationDone()

	debugDumpers := make(map[string]debugserver.Dumper)
	debugDumpers["repos"] = updateScheduler
	debugserverEndpoints.repoUpdaterStateEndpoint = repoUpdaterStatsHandler(debugDumpers)
	debugserverEndpoints.listAuthzProvidersEndpoint = listAuthzProvidersHandler()
	debugserverEndpoints.gitserverReposStatusEndpoint = gitserverReposStatusHandler(db)
	debugserverEndpoints.manualPurgeEndpoint = manualPurgeHandler(db)

	// We mark the service as ready now AFTER assigning the additional endpoints in
	// the debugserver constructed at the top of this function. This ensures we don't
	// have a race between becoming ready and a debugserver request failing directly
	// after being unblocked.
	ready()

	goroutine.MonitorBackgroundRoutines(ctx, routines...)

	return nil
}

func getDB(observationCtx *observation.Context) (database.DB, error) {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "repo-updater")
	if err != nil {
		return nil, errors.Wrap(err, "initializing database store")
	}
	return database.NewDB(observationCtx.Logger, sqlDB), nil
}

func makeHTTPServer(logger log.Logger, server *repoupdater.Server) goroutine.BackgroundRoutine {
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}

	addr := net.JoinHostPort(host, port)
	logger.Info("listening", log.String("addr", addr))

	m := repoupdater.NewHandlerMetrics()
	m.MustRegister(prometheus.DefaultRegisterer)
	handler := repoupdater.ObservedHandler(
		logger,
		m,
		otel.GetTracerProvider(),
	)(server.Handler())
	grpcServer := grpc.NewServer(defaults.ServerOptions(logger)...)
	serviceServer := &repoupdater.RepoUpdaterServiceServer{
		Server: server,
	}
	proto.RegisterRepoUpdaterServiceServer(grpcServer, serviceServer)
	reflection.Register(grpcServer)
	handler = internalgrpc.MultiplexHandlers(grpcServer, handler)

	// NOTE: Internal actor is required to have full visibility of the repo table
	// 	(i.e. bypass repository authorization).
	authzBypass := func(f http.Handler) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			r = r.WithContext(actor.WithInternalActor(r.Context()))
			f.ServeHTTP(w, r)
		}
	}

	return httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler: instrumentation.HTTPMiddleware(
			"",
			trace.HTTPMiddleware(logger, authzBypass(handler), conf.DefaultClient()),
		),
	})
}

func gitserverReposStatusHandler(db database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		repo := r.FormValue("repo")
		if repo == "" {
			http.Error(w, "missing 'repo' param", http.StatusBadRequest)
			return
		}

		status, err := db.GitserverRepos().GetByName(r.Context(), api.RepoName(repo))
		if err != nil {
			http.Error(w, fmt.Sprintf("fetching repository status: %q", err), http.StatusInternalServerError)
			return
		}

		resp, err := json.MarshalIndent(status, "", "  ")
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to marshal status: %q", err.Error()), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
	}
}

func manualPurgeHandler(db database.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid limit: %v", err), http.StatusBadRequest)
			return
		}
		if limit <= 0 {
			http.Error(w, "limit must be greater than 0", http.StatusBadRequest)
			return
		}
		if limit > 10000 {
			http.Error(w, "limit must be less than 10000", http.StatusBadRequest)
			return
		}
		perSecond := 1.0 // Default value
		perSecondParam := r.FormValue("perSecond")
		if perSecondParam != "" {
			perSecond, err = strconv.ParseFloat(perSecondParam, 64)
			if err != nil {
				http.Error(w, fmt.Sprintf("invalid per second rate limit: %v", err), http.StatusBadRequest)
				return
			}
			// Set a sane lower bound
			if perSecond <= 0.1 {
				http.Error(w, fmt.Sprintf("invalid per second rate limit. Must be > 0.1, got %f", perSecond), http.StatusBadRequest)
				return
			}
		}
		err = repos.PurgeOldestRepos(log.Scoped("PurgeOldestRepos"), db, limit, perSecond)
		if err != nil {
			http.Error(w, fmt.Sprintf("starting manual purge: %v", err), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(fmt.Sprintf("manual purge started with limit of %d and rate of %f", limit, perSecond)))
	}
}

func listAuthzProvidersHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type providerInfo struct {
			ServiceType        string `json:"service_type"`
			ServiceID          string `json:"service_id"`
			ExternalServiceURL string `json:"external_service_url"`
		}

		_, providers := ossAuthz.GetProviders()
		infos := make([]providerInfo, len(providers))
		for i, p := range providers {
			_, id := extsvc.DecodeURN(p.URN())

			// Note that the ID marshalling below replicates code found in `graphqlbackend`.
			// We cannot import that package's code into this one (see /dev/check/go-dbconn-import.sh).
			infos[i] = providerInfo{
				ServiceType:        p.ServiceType(),
				ServiceID:          p.ServiceID(),
				ExternalServiceURL: fmt.Sprintf("%s/site-admin/external-services/%s", globals.ExternalURL(), relay.MarshalID("ExternalService", id)),
			}
		}

		resp, err := json.MarshalIndent(infos, "", "  ")
		if err != nil {
			http.Error(w, "failed to marshal infos: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(resp)
	}
}

func repoUpdaterStatsHandler(debugDumpers map[string]debugserver.Dumper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		wantDumper := r.URL.Query().Get("dumper")
		wantFormat := r.URL.Query().Get("format")

		// Showing the HTML version of repository syncing schedule as the default,
		// also the only dumper that supports rendering the HTML version.
		if (wantDumper == "" || wantDumper == "repos") && wantFormat != "json" {
			reposDumper, ok := debugDumpers["repos"].(*scheduler.UpdateScheduler)
			if !ok {
				http.Error(w, "No debug dumper for repos found", http.StatusInternalServerError)
				return
			}

			// This case also applies for defaultOffer. Note that this is preferred
			// over e.g. a 406 status code, according to the MDN:
			// https://developer.mozilla.org/en-US/docs/Web/HTTP/Status/406
			tmpl := template.New("state.html").Funcs(template.FuncMap{
				"truncateDuration": func(d time.Duration) time.Duration {
					return d.Truncate(time.Second)
				},
			})
			template.Must(tmpl.Parse(stateHTMLTemplate))
			err := tmpl.Execute(w, reposDumper.DebugDump(r.Context()))
			if err != nil {
				http.Error(w, "Failed to render template: "+err.Error(), http.StatusInternalServerError)
				return
			}
			return
		}

		var dumps []any
		for name, dumper := range debugDumpers {
			if wantDumper != "" && wantDumper != name {
				continue
			}
			dumps = append(dumps, dumper.DebugDump(r.Context()))
		}

		p, err := json.MarshalIndent(dumps, "", "  ")
		if err != nil {
			http.Error(w, "Failed to marshal dumps: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(p)
	}
}

func watchSyncer(
	ctx context.Context,
	logger log.Logger,
	syncer *repos.Syncer,
	sched *scheduler.UpdateScheduler,
	changesetSyncer syncer.UnarchivedChangesetSyncRegistry,
) {
	logger.Debug("started new repo syncer updates scheduler relay thread")

	for {
		select {
		case <-ctx.Done():
			return
		case diff := <-syncer.Synced:
			if !conf.Get().DisableAutoGitUpdates {
				sched.UpdateFromDiff(diff)
			}

			// Similarly, changesetSyncer is only available in enterprise mode.
			if changesetSyncer != nil {
				repositories := diff.Modified.ReposModified(types.RepoModifiedArchived)
				if len(repositories) > 0 {
					if err := changesetSyncer.EnqueueChangesetSyncsForRepos(ctx, repositories.IDs()); err != nil {
						logger.Warn("error enqueuing changeset syncs for archived and unarchived repos", log.Error(err))
					}
				}
			}
		}
	}
}

// newUnclonedReposManager creates a background routine that will periodically list
// the uncloned repositories on gitserver and update the scheduler with the list.
// It also ensures that if any of our indexable repos are missing from the cloned
// list they will be added for cloning ASAP.
func newUnclonedReposManager(ctx context.Context, logger log.Logger, isSourcegraphDotCom bool, sched *scheduler.UpdateScheduler, store repos.Store) goroutine.BackgroundRoutine {
	return goroutine.NewPeriodicGoroutine(
		actor.WithInternalActor(ctx),
		goroutine.HandlerFunc(func(ctx context.Context) error {
			// Don't modify the scheduler if we're not performing auto updates.
			if conf.Get().DisableAutoGitUpdates {
				return nil
			}

			baseRepoStore := database.ReposWith(logger, store)

			if isSourcegraphDotCom {
				// Fetch ALL indexable repos that are NOT cloned so that we can add them to the
				// scheduler.
				opts := database.ListSourcegraphDotComIndexableReposOptions{
					CloneStatus: types.CloneStatusNotCloned,
				}
				indexable, err := baseRepoStore.ListSourcegraphDotComIndexableRepos(ctx, opts)
				if err != nil {
					return errors.Wrap(err, "listing indexable repos")
				}
				// Ensure that uncloned indexable repos are known to the scheduler
				sched.EnsureScheduled(indexable)
			}

			// Next, move any repos managed by the scheduler that are uncloned to the front
			// of the queue.
			managed := sched.ListRepoIDs()

			uncloned, err := baseRepoStore.ListMinimalRepos(ctx, database.ReposListOptions{IDs: managed, NoCloned: true})
			if err != nil {
				return errors.Wrap(err, "failed to fetch list of uncloned repositories")
			}

			sched.PrioritiseUncloned(uncloned)

			return nil
		}),
		goroutine.WithName("repo-updater.uncloned-repo-manager"),
		goroutine.WithDescription("periodically lists uncloned repos and schedules them as high priority in the repo updater update queue"),
		goroutine.WithInterval(30*time.Second),
	)
}

// TODO: This might clash with what osscmd.Main does.
// watchAuthzProviders updates authz providers if config changes.
func watchAuthzProviders(ctx context.Context, db database.DB) {
	globals.WatchPermissionsUserMapping()
	go func() {
		t := time.NewTicker(providers.RefreshInterval())
		for range t.C {
			allowAccessByDefault, authzProviders, _, _, _ := providers.ProvidersFromConfig(
				ctx,
				conf.Get(),
				db,
			)
			ossAuthz.SetProviders(allowAccessByDefault, authzProviders)
		}
	}()
}
