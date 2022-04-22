// Command searcher is a simple service which exposes an API to text search a
// repo at a specific commit. See the searcher package for more information.
package main

import (
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/inconshreveable/log15"

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
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/version"
	sglog "github.com/sourcegraph/sourcegraph/lib/log"
)

var (
	cacheDir    = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	cacheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "100000", "maximum size of the on disk cache in megabytes")
)

const port = "3181"

func ensureFrontendDB() database.DB {
	dsn := conf.GetServiceConnectionValueAndRestartOnChange(func(serviceConnections conftypes.ServiceConnections) string {
		return serviceConnections.PostgresDSN
	})
	sqlDB, err := connections.EnsureNewFrontendDB(dsn, "searcher", &observation.TestContext)
	if err != nil {
		log.Fatalf("Failed to connect to frontend database: %s", err)
	}
	return database.NewDB(sqlDB)
}

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	conf.Init()
	logging.Init()
	sglog.Init(sglog.Resource{
		Name:    env.MyName,
		Version: version.Version(),
	})
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()
	profiler.Init()

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	var cacheSizeBytes int64
	if i, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		log.Fatalf("invalid int %q for SEARCHER_CACHE_SIZE_MB: %s", cacheSizeMB, err)
	} else {
		cacheSizeBytes = i * 1000 * 1000
	}

	db := ensureFrontendDB()
	git := gitserver.NewClient(db)

	service := &search.Service{
		Store: &search.Store{
			FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID) (io.ReadCloser, error) {
				return git.Archive(ctx, repo, gitserver.ArchiveOptions{
					Treeish: string(commit),
					Format:  "tar",
				})
			},
			FetchTarPaths: func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
				pathspecs := make([]gitserver.Pathspec, len(paths))
				for i, p := range paths {
					pathspecs[i] = gitserver.PathspecLiteral(p)
				}
				return git.Archive(ctx, repo, gitserver.ArchiveOptions{
					Treeish:   string(commit),
					Format:    "tar",
					Pathspecs: pathspecs,
				})
			},
			FilterTar:         search.NewFilter,
			Path:              filepath.Join(cacheDir, "searcher-archives"),
			MaxCacheSizeBytes: cacheSizeBytes,
			DB:                db,
		},
		Log: sglog.Scoped("service", "the searcher service"),
	}
	service.Store.Start()

	// Set up handler middleware
	handler := actor.HTTPMiddleware(service)
	handler = trace.HTTPMiddleware(handler, conf.DefaultClient())
	handler = ot.HTTPMiddleware(handler)

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	server := &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Addr:         addr,
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

	go func() {
		log15.Info("searcher: listening", "addr", server.Addr)
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Listen for shutdown signals. When we receive one attempt to clean up,
	// but do an insta-shutdown if we receive more than one signal.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP, syscall.SIGTERM)

	// Once we receive one of the signals from above, continues with the shutdown
	// process.
	<-c
	go func() {
		// If a second signal is received, exit immediately.
		<-c
		os.Exit(0)
	}()

	// Wait for at most for the configured shutdown timeout.
	ctx, cancel := context.WithTimeout(context.Background(), goroutine.GracefulShutdownTimeout)
	defer cancel()
	// Stop accepting requests.
	if err := server.Shutdown(ctx); err != nil {
		log15.Error("shutting down http server", "error", err)
	}
}
