// Package shared is the shared main entrypoint for searcher, a simple service which exposes an API
// to text search a repo at a specific commit. See the searcher package for more information.
package shared

import (
	"context"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/keegancsmith/tmpfriend"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/conf/conftypes"
	"github.com/sourcegraph/sourcegraph/internal/database"
	connections "github.com/sourcegraph/sourcegraph/internal/database/connections/live"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	sharedsearch "github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	cacheDir    = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	cacheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "100000", "maximum size of the on disk cache in megabytes")

	maxTotalPathsLengthRaw = env.Get("MAX_TOTAL_PATHS_LENGTH", "100000", "maximum sum of lengths of all paths in a single call to git archive")
)

const port = "3181"

func frontendDB(observationCtx *observation.Context) (database.DB, error) {
	logger := log.Scoped("frontendDB", "")
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(observationCtx, dsn, "searcher")
	if err != nil {
		return nil, err
	}
	return database.NewDB(logger, sqlDB), nil
}

func shutdownOnSignal(ctx context.Context, server *http.Server) error {
	// Listen for shutdown signals. When we receive one attempt to clean up,
	// but do an insta-shutdown if we receive more than one signal.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	// Once we receive one of the signals from above, continues with the shutdown
	// process.
	select {
	case <-c:
	case <-ctx.Done(): // still call shutdown below
	}

	go func() {
		// If a second signal is received, exit immediately.
		<-c
		os.Exit(1)
	}()

	// Wait for at most for the configured shutdown timeout.
	ctx, cancel := context.WithTimeout(ctx, goroutine.GracefulShutdownTimeout)
	defer cancel()
	// Stop accepting requests.
	return server.Shutdown(ctx)
}

// setupTmpDir sets up a temporary directory on the same volume as the
// cacheDir.
//
// Structural search relies on temporary files created from zoekt responses.
// Additionally we shell out to programs that may or may not need a temporary
// directory.
//
// search.Store will also take into account the files in tmp when deciding on
// evicting items due to disk pressure. It won't delete those files unless
// they are zip files. In the case of comby the files are temporary so them
// being deleted while read by comby is fine since it will maintain an open
// FD.
func setupTmpDir() error {
	tmpRoot := filepath.Join(cacheDir, ".searcher.tmp")
	if err := os.MkdirAll(tmpRoot, 0o755); err != nil {
		return err
	}
	if !tmpfriend.IsTmpFriendDir(tmpRoot) {
		_, err := tmpfriend.RootTempDir(tmpRoot)
		return err
	}
	return nil
}

func Start(ctx context.Context, observationCtx *observation.Context) error {
	logger := observationCtx.Logger

	// TODO(sqs):
	//
	// 1. Don't unconditionally spawn a debugserver here, since it might conflict with other
	//    services' debugservers. (Note: to allow it to run and not crash on startup, I've disabled
	//    it with `if os.Getenv("DEPLOY_TYPE") != "single-program" { ... }`.
	//
	// 2. Use the standard goroutine package's monitoring to start/stop debugserver and the other
	//    goroutines. Return a list of goroutine.BackgroundProcess interfaces.

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	if os.Getenv("DEPLOY_TYPE") != "single-program" {
		go debugserver.NewServerRoutine(ready).Start()
	}

	var cacheSizeBytes int64
	if i, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		return errors.Wrapf(err, "invalid int %q for SEARCHER_CACHE_SIZE_MB", cacheSizeMB)
	} else {
		cacheSizeBytes = i * 1000 * 1000
	}

	maxTotalPathsLength, err := strconv.Atoi(maxTotalPathsLengthRaw)
	if err != nil {
		return errors.Wrapf(err, "invalid int %q for MAX_TOTAL_PATHS_LENGTH", maxTotalPathsLengthRaw)
	}

	if err := setupTmpDir(); err != nil {
		return errors.Wrap(err, "failed to setup TMPDIR")
	}

	// Explicitly don't scope Store logger under the parent logger
	storeObservationCtx := observation.NewContext(log.Scoped("Store", "searcher archives store"))

	db, err := frontendDB(observation.NewContext(log.Scoped("db", "server frontend db")))
	if err != nil {
		return errors.Wrap(err, "failed to connect to frontend database")
	}
	git := gitserver.NewClient(db)

	service := &search.Service{
		Store: &search.Store{
			FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID) (io.ReadCloser, error) {
				// We pass in a nil sub-repo permissions checker and an internal actor here since
				// searcher needs access to all data in the archive.
				ctx = actor.WithInternalActor(ctx)
				return git.ArchiveReader(ctx, nil, repo, gitserver.ArchiveOptions{
					Treeish: string(commit),
					Format:  gitserver.ArchiveFormatTar,
				})
			},
			FetchTarPaths: func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
				pathspecs := make([]gitdomain.Pathspec, len(paths))
				for i, p := range paths {
					pathspecs[i] = gitdomain.PathspecLiteral(p)
				}
				// We pass in a nil sub-repo permissions checker and an internal actor here since
				// searcher needs access to all data in the archive.
				ctx = actor.WithInternalActor(ctx)
				return git.ArchiveReader(ctx, nil, repo, gitserver.ArchiveOptions{
					Treeish:   string(commit),
					Format:    gitserver.ArchiveFormatTar,
					Pathspecs: pathspecs,
				})
			},
			FilterTar:         search.NewFilter,
			Path:              filepath.Join(cacheDir, "searcher-archives"),
			MaxCacheSizeBytes: cacheSizeBytes,
			Log:               storeObservationCtx.Logger,
			ObservationCtx:    storeObservationCtx,
			DB:                db,
		},

		Indexed: sharedsearch.Indexed(),

		GitDiffSymbols: func(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) ([]byte, error) {
			// As this is an internal service call, we need an internal actor.
			ctx = actor.WithInternalActor(ctx)
			return git.DiffSymbols(ctx, repo, commitA, commitB)
		},
		MaxTotalPathsLength: maxTotalPathsLength,

		Log: logger,
	}
	service.Store.Start()

	// Set up handler middleware
	handler := actor.HTTPMiddleware(logger, service)
	handler = trace.HTTPMiddleware(logger, handler, conf.DefaultClient())
	handler = instrumentation.HTTPMiddleware("", handler)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	g, ctx := errgroup.WithContext(ctx)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	server := &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Addr:         addr,
		BaseContext: func(_ net.Listener) context.Context {
			return ctx
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// For cluster liveness and readiness probes
			if r.URL.Path == "/healthz" {
				w.WriteHeader(200)
				_, _ = w.Write([]byte("ok"))
				return
			}
			handler.ServeHTTP(w, r)
		}),
	}

	// Listen
	g.Go(func() error {
		logger.Info("listening", log.String("addr", server.Addr))
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return err
		}
		return nil
	})

	// Shutdown
	g.Go(func() error {
		return shutdownOnSignal(ctx, server)
	})

	return g.Wait()
}
