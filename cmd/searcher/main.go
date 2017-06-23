// search is a simple service which exposes an API to text search a repo at
// a specific commit. See the searcher package for more information.
package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/searcher/search"
	sourcegraph "sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/vcs"
)

var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")
var cacheDir = env.Get("CACHE_DIR", "/tmp", "directory to store cached archives.")
var cacheSizeMB = env.Get("SEARCHER_CACHE_SIZE_MB", "0", "maximum size of the on disk cache in megabytes")

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	tracer.Init("searcher")
	gitserver.DefaultClient.NoCreds = true

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
	}

	var cacheSizeBytes int64
	if i, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		log.Fatalf("invalid int %q for SEARCHER_CACHE_SIZE_MB: %s", cacheSizeMB, err)
	} else {
		cacheSizeBytes = i * 1000 * 1000
	}

	service := &search.Service{
		Store: &search.Store{
			FetchTar:          fetchTar,
			Path:              filepath.Join(cacheDir, "searcher-archives"),
			MaxCacheSizeBytes: cacheSizeBytes,
		},
		RequestLog: log.New(os.Stderr, "", 0),
	}
	service.Store.Start()
	handler := nethttp.Middleware(opentracing.GlobalTracer(), service)

	addr := ":3181"
	server := &http.Server{Addr: addr, Handler: handler}
	go shutdownOnSIGINT(server)

	log.Println("listening on :3181")
	err := server.ListenAndServe()
	if err != http.ErrServerClosed {
		log.Fatal(err)
	}
}

func shutdownOnSIGINT(s *http.Server) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err := s.Shutdown(ctx)
	if err != nil {
		log.Fatal("graceful server shutdown failed, will exit:", err)
	}
}

func fetchTar(ctx context.Context, repo, commit string) (r io.ReadCloser, err error) {
	// gitcmd.Repository.Archive returns a zip file read into
	// memory. However, we do not need to read into memory and we want a
	// tar, so we directly run the gitserver Command.
	span, ctx := opentracing.StartSpanFromContext(ctx, "OpenTar")
	ext.Component.Set(span, "git")
	span.SetTag("URL", repo)
	span.SetTag("Commit", commit)
	defer func() {
		if err != nil {
			ext.Error.Set(span, true)
			span.SetTag("err", err)
		}
		span.Finish()
	}()

	if strings.HasPrefix(commit, "-") {
		return nil, badRequestError{("invalid git revision spec (begins with '-')")}
	}

	cmd := gitserver.DefaultClient.Command("git", "archive", "--format=tar", commit)
	cmd.Repo = &sourcegraph.Repo{URI: repo}
	cmd.EnsureRevision = commit
	r, err = gitserver.StdoutReader(ctx, cmd)
	if err != nil {
		if vcs.IsRepoNotExist(err) || err == vcs.ErrRevisionNotFound {
			err = badRequestError{err.Error()}
		}
		return nil, err
	}
	return r, nil
}

type badRequestError struct{ msg string }

func (e badRequestError) Error() string    { return e.msg }
func (e badRequestError) BadRequest() bool { return true }
