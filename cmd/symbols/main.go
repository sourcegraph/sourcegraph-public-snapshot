// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/parser"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/sqlite"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/symbols"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/diskcache"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
	"github.com/sourcegraph/sourcegraph/internal/httpserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/profiler"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const addr = ":3184"

func main() {
	config.Load()

	// Set up Google Cloud Profiler when running in Cloud
	if err := profiler.Init(); err != nil {
		log.Fatalf("Failed to start profiler: %v", err)
	}

	env.Lock()
	env.HandleHelpFlag()
	conf.Init()
	logging.Init()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	if err := config.Validate(); err != nil {
		log.Fatalf("Failed to load configuration: %s", err)
	}

	if config.sanityCheck {
		fmt.Print("Running sanity check...")
		if err := sqlite.SanityCheck(); err != nil {
			fmt.Println("failed ❌", err)
			os.Exit(1)
		}

		fmt.Println("passed ✅")
		os.Exit(0)
	}

	// Start debug server
	ready := make(chan struct{})
	go debugserver.NewServerRoutine(ready).Start()

	cache := &diskcache.Store{
		Dir:               config.cacheDir,
		Component:         "symbols",
		BackgroundTimeout: 20 * time.Minute,
	}

	parserPool, err := parser.NewParserPool(parser.NewParser, config.ctagsProcesses)
	if err != nil {
		log.Fatalf("Failed to parser pool: %s", err)
	}

	server := httpserver.NewFromAddr(addr, &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Handler:      ot.Middleware(trace.HTTPTraceMiddleware(symbols.NewHandler(&gitserverClient{}, cache, parserPool, 15))),
	})

	evictionDuration := time.Second * 10
	cacheSizeBytes := int64(config.cacheSizeMB) * 1000 * 1000
	cacheEvicter := sqlite.NewCacheEvicter(evictionDuration, cache, cacheSizeBytes)

	// Mark health server as ready and go!
	close(ready)
	goroutine.MonitorBackgroundRoutines(context.Background(), server, cacheEvicter)
}

type gitserverClient struct{}

func (c *gitserverClient) FetchTar(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
	return gitserver.DefaultClient.Archive(ctx, repo, gitserver.ArchiveOptions{Treeish: string(commit), Format: "tar", Paths: paths})
}

func (c *gitserverClient) GitDiff(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) (*parser.Changes, error) {
	command := gitserver.DefaultClient.Command("git", "diff", "-z", "--name-status", "--no-renames", string(commitA), string(commitB))
	command.Repo = repo

	output, err := command.Output(ctx)
	if err != nil {
		return nil, err
	}

	// The output is a a repeated sequence of:
	//
	//     <status> NUL <path> NUL
	//
	// where NUL is the 0 byte.
	//
	// Example:
	//
	//     M NUL cmd/symbols/internal/symbols/fetch.go NUL

	changes := parser.Changes{}
	slices := bytes.Split(output, []byte{0})
	for i := 0; i < len(slices)-1; i += 2 {
		statusIdx := i
		fileIdx := i + 1

		if len(slices[statusIdx]) == 0 {
			return nil, fmt.Errorf("unrecognized git diff output (from repo %q, commitA %q, commitB %q): status was empty at index %d", repo, commitA, commitB, i)
		}

		status := slices[statusIdx][0]
		path := string(slices[fileIdx])

		switch status {
		case 'A':
			changes.Added = append(changes.Added, path)
		case 'M':
			changes.Modified = append(changes.Modified, path)
		case 'D':
			changes.Deleted = append(changes.Deleted, path)
		}
	}

	return &changes, nil
}
