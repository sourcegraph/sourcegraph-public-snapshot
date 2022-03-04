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

	"github.com/sourcegraph/sourcegraph/cmd/searcher/search"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/store"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

var (
	cacheDir    = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
	cacheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "100000", "maximum size of the on disk cache in megabytes")
)

const port = "3181"

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	conf.Init()
	logging.Init()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	if err := profiler.Init(); err != nil {
		log.Fatalf("failed to start Google Cloud profiler: %s", err)
	}

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

	service := &search.Service{
		Store: &store.Store{
			FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID) (io.ReadCloser, error) {
				return git.ArchiveReader(ctx, repo, gitserver.ArchiveOptions{Treeish: string(commit), Format: git.ArchiveFormatTar})
			},
			FilterTar:         search.NewFilter,
			Path:              filepath.Join(cacheDir, "searcher-archives"),
			MaxCacheSizeBytes: cacheSizeBytes,
		},
		Log: log15.Root(),
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
