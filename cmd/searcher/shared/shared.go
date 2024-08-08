// Package shared is the shared main entrypoint for searcher, a simple service which exposes an API
// to text search a repo at a specific commit. See the searcher package for more information.
package shared

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/keegancsmith/tmpfriend"
	"github.com/sourcegraph/log"
	"google.golang.org/grpc"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	internalgrpc "github.com/sourcegraph/sourcegraph/internal/grpc"
	"github.com/sourcegraph/sourcegraph/internal/grpc/defaults"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/instrumentation"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/requestclient"
	"github.com/sourcegraph/sourcegraph/internal/requestinteraction"
	sharedsearch "github.com/sourcegraph/sourcegraph/internal/search"
	proto "github.com/sourcegraph/sourcegraph/internal/searcher/v1"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/tenant"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
func setupTmpDir(cacheDir string) error {
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

func Start(ctx context.Context, observationCtx *observation.Context, ready service.ReadyFunc, cfg *Config) error {
	logger := observationCtx.Logger

	// Load and validate configuration.
	if err := cfg.Validate(); err != nil {
		return errors.Wrap(err, "failed to validate configuration")
	}

	if err := setupTmpDir(cfg.CacheDir); err != nil {
		return errors.Wrap(err, "failed to setup TMPDIR")
	}

	git := gitserver.NewClient("searcher")

	// Explicitly don't scope Store logger under the parent logger
	storeObservationCtx := observation.NewContext(log.Scoped("Store"))
	store := &search.Store{
		FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
			return git.ArchiveReader(ctx, repo, gitserver.ArchiveOptions{
				Treeish: string(commit),
				Format:  gitserver.ArchiveFormatTar,
				Paths:   paths,
			})
		},
		FilterTar:         search.NewFilterFactory(git),
		Path:              filepath.Join(cfg.CacheDir, "searcher-archives"),
		MaxCacheSizeBytes: int64(cfg.CacheSizeMB * 1000 * 1000),
		BackgroundTimeout: cfg.BackgroundTimeout,
		Logger:            storeObservationCtx.Logger,
		ObservationCtx:    storeObservationCtx,
	}
	store.Start()

	sService := &search.Service{
		Store:   store,
		Indexed: sharedsearch.Indexed(),
		GitChangedFiles: func(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) (gitserver.ChangedFilesIterator, error) {
			return git.ChangedFiles(ctx, repo, string(commitA), string(commitB))
		},
		MaxTotalPathsLength: cfg.MaxTotalGitArchivePathsLength,
		Logger:              logger,
		DisableHybridSearch: cfg.DisableHybridSearch,
	}

	grpcServer := defaults.NewServer(logger)
	proto.RegisterSearcherServiceServer(grpcServer, search.NewGRPCServer(sService, cfg.ExhaustiveRequestLoggingEnabled))

	ready()

	logger.Info("searcher: listening", log.String("addr", cfg.ListenAddress))

	return goroutine.MonitorBackgroundRoutines(ctx, makeHTTPServer(logger, grpcServer, cfg.ListenAddress))
}

// makeHTTPServer creates a new *http.Server for the searcher endpoints and registers
// it with methods on the given server. It multiplexes HTTP requests and gRPC requests
// from a single port.
func makeHTTPServer(logger log.Logger, grpcServer *grpc.Server, listenAddress string) goroutine.BackgroundRoutine {
	// TODO: This should be removed, and gRPC only should be served instead.
	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// For cluster liveness and readiness probes
		if r.URL.Path == "/healthz" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte("ok"))
			return
		}
		http.NotFoundHandler().ServeHTTP(w, r)
	})
	handler = actor.HTTPMiddleware(logger, handler)
	handler = tenant.InternalHTTPMiddleware(logger, handler)
	handler = requestclient.InternalHTTPMiddleware(handler)
	handler = requestinteraction.HTTPMiddleware(handler)
	handler = trace.HTTPMiddleware(logger, handler)
	handler = instrumentation.HTTPMiddleware("", handler)
	handler = internalgrpc.MultiplexHandlers(grpcServer, handler)

	return httpserver.NewFromAddr(listenAddress, &http.Server{
		Handler: handler,
	})
}
