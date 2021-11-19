// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/symbols"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/debugserver"
	"github.com/sourcegraph/sourcegraph/internal/env"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/logging"
	"github.com/sourcegraph/sourcegraph/internal/sentry"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/trace/ot"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
)

const port = "3184"

func main() {
	var (
		cacheDir       = env.Get("CACHE_DIR", "/tmp/symbols-cache", "directory to store cached symbols")
		cacheSizeMB    = env.Get("SYMBOLS_CACHE_SIZE_MB", "100000", "maximum size of the disk cache in megabytes")
		ctagsProcesses = env.Get("CTAGS_PROCESSES", strconv.Itoa(runtime.GOMAXPROCS(0)), "number of ctags child processes to run")
		sanityCheck    = env.Get("SANITY_CHECK", "false", "check that go-sqlite3 works then exit 0 if it's ok or 1 if not")
	)

	if sanityCheck == "true" {
		fmt.Print("Running sanity check...")
		if err := symbols.SanityCheck(); err != nil {
			fmt.Println("failed ❌", err)
			os.Exit(1)
		}

		fmt.Println("passed ✅")
		os.Exit(0)
	}

	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	conf.Init()
	logging.Init()
	tracer.Init(conf.DefaultClient())
	sentry.Init(conf.DefaultClient())
	trace.Init()

	// Ready immediately
	ready := make(chan struct{})
	close(ready)
	go debugserver.NewServerRoutine(ready).Start()

	service := symbols.Service{
		FetchTar: func(ctx context.Context, repo api.RepoName, commit api.CommitID, paths []string) (io.ReadCloser, error) {
			return gitserver.DefaultClient.Archive(ctx, repo, gitserver.ArchiveOptions{Treeish: string(commit), Format: "tar", Paths: paths})
		},
		GitDiff: func(ctx context.Context, repo api.RepoName, commitA, commitB api.CommitID) (*symbols.Changes, error) {
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

			changes := symbols.NewChanges()
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
		},
		NewParser: symbols.NewParser,
		Path:      cacheDir,
	}
	if mb, err := strconv.ParseInt(cacheSizeMB, 10, 64); err != nil {
		log.Fatalf("Invalid SYMBOLS_CACHE_SIZE_MB: %s", err)
	} else {
		service.MaxCacheSizeBytes = mb * 1000 * 1000
	}
	var err error
	service.NumParserProcesses, err = strconv.Atoi(ctagsProcesses)
	if err != nil {
		log.Fatalf("Invalid CTAGS_PROCESSES: %s", err)
	}
	if err := service.Start(); err != nil {
		log.Fatalln("Start:", err)
	}

	handler := ot.Middleware(trace.HTTPTraceMiddleware(service.Handler()))

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	server := &http.Server{
		ReadTimeout:  75 * time.Second,
		WriteTimeout: 10 * time.Minute,
		Addr:         addr,
		Handler:      handler,
	}
	go shutdownOnSIGINT(server)

	log15.Info("symbols: listening", "addr", addr)
	err = server.ListenAndServe()
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
