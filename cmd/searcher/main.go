// Command searcher is a simple service which exposes an API to text search a
// repo at a specific commit. See the searcher package for more information.
package main

import (
	"context"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"golang.org/x/sync/errgroup"

	"github.com/getsentry/sentry-go"
	"github.com/keegancsmith/tmpfriend"
	"github.com/prometheus/client_golang/prometheus"
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
	"github.com/sourcegraph/sourcegraph/internal/hostname"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var (
	cacheDir    = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	cacheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "100000", "maximum size of the on disk cache in megabytes")
)

const port = "3181"

func frontendDB() (database.DB, error) {
	logger := log.Scoped("frontendDB", "")
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(dsn, "searcher", &observation.TestContext)
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
	if err := os.MkdirAll(tmpRoot, 0755); err != nil {
		return err
	}
	if !tmpfriend.IsTmpFriendDir(tmpRoot) {
		_, err := tmpfriend.RootTempDir(tmpRoot)
		return err
	}
	return nil
}

func run(logger log.Logger) error {
	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	var cacheSizeBytes int64
	if i, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		return errors.Wrapf(err, "invalid int %q for SEARCHER_CACHE_SIZE_MB", cacheSizeMB)
	} else {
		cacheSizeBytes = i * 1000 * 1000
	}

	if err := setupTmpDir(); err != nil {
		return errors.Wrap(err, "failed to setup TMPDIR")
	}

	storeObservationContext := &observation.Context{
		// Explicitly don't scope Store logger under the parent logger
		Logger:     log.Scoped("Store", "searcher archives store"),
		Tracer:     &trace.Tracer{TracerProvider: otel.GetTracerProvider()},
		Registerer: prometheus.DefaultRegisterer,
	}

	db, err := frontendDB()
	if err != nil {
		return errors.Wrap(err, "failed to connect to frontend database")
	}
	git := gitserver.NewClient(db)

	service := &search.Service{
		Store: &search.Store{
			FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID) (io.ReadCloser, error) {
				// We pass in a nil sub-repo permissions checker here since searcher needs access
				// to all data in the archive
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
				// We pass in a nil sub-repo permissions checker here since searcher needs access
				// to all data in the archive
				return git.ArchiveReader(ctx, nil, repo, gitserver.ArchiveOptions{
					Treeish:   string(commit),
					Format:    gitserver.ArchiveFormatTar,
					Pathspecs: pathspecs,
				})
			},
			FilterTar:          search.NewFilter,
			Path:               filepath.Join(cacheDir, "searcher-archives"),
			MaxCacheSizeBytes:  cacheSizeBytes,
			Log:                storeObservationContext.Logger,
			ObservationContext: storeObservationContext,
			DB:                 db,
		},
		GitDiffSymbols: git.DiffSymbols,
		Log:            logger,
	}
	service.Store.Start()

	// Set up handler middleware
	handler := actor.HTTPMiddleware(service)
	handler = trace.HTTPMiddleware(logger, handler, conf.DefaultClient())
	handler = instrumentation.HTTPMiddleware("", handler)

	ctx, cancel := context.WithCancel(context.Background())
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

func main() {
	env.Lock()
	env.HandleHelpFlag()
	stdlog.SetFlags(0)
	logging.Init()
	liblog := log.Init(log.Resource{
		Name:       env.MyName,
		Version:    version.Version(),
		InstanceID: hostname.Get(),
	}, log.NewSentrySinkWith(
		log.SentrySink{
			ClientOptions: sentry.ClientOptions{SampleRate: 0.2},
		},
	)) // Experimental: DevX is observing how sampling affects the errors signal
	defer liblog.Sync()

	conf.Init()
	go conf.Watch(liblog.Update(conf.GetLogSinks))
	tracer.Init(log.Scoped("tracer", "internal tracer package"), conf.DefaultClient())
	trace.Init()
	profiler.Init()

	logger := log.Scoped("server", "the searcher service")

	err := run(logger)
	if err != nil {
		logger.Fatal("searcher failed", log.Error(err))
	}
}
