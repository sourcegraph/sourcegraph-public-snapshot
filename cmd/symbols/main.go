// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"context"
	"database/sql"
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

	sqlite3 "github.com/mattn/go-sqlite3"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/pkg/ctags"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/internal/symbols"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/gitserver"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
	"github.com/sourcegraph/sourcegraph/pkg/vcs/git"
)

var (
	cacheDir       = env.Get("CACHE_DIR", "/tmp/symbols-cache", "directory to store cached symbols")
	cacheSizeMB    = env.Get("SYMBOLS_CACHE_SIZE_MB", "100000", "maximum size of the disk cache in megabytes")
	ctagsProcesses = env.Get("CTAGS_PROCESSES", strconv.Itoa(runtime.NumCPU()), "number of ctags child processes to run")
	ctagsCommand   = env.Get("CTAGS_COMMAND", "universal-ctags", "ctags command (should point to universal-ctags executable compiled with JSON and seccomp support)")
	libSqlite3Pcre = env.Get("LIBSQLITE3_PCRE", "", "path to the libsqlite3-pcre library")
)

const port = "3184"

func main() {
	env.Lock()
	env.HandleHelpFlag()
	log.SetFlags(0)
	tracer.Init()

	if libSqlite3Pcre == "" {
		env.PrintHelp()
		panic("Can't find the libsqlite3-pcre library because LIBSQLITE3_PCRE was not set.")
	}
	sql.Register("sqlite3_with_pcre", &sqlite3.SQLiteDriver{Extensions: []string{libSqlite3Pcre}})

	go debugserver.Start()

	service := symbols.Service{
		FetchTar: func(ctx context.Context, repo gitserver.Repo, commit api.CommitID) (io.ReadCloser, error) {
			return git.Archive(ctx, repo, git.ArchiveOptions{Treeish: string(commit), Format: "tar"})
		},
		NewParser: func() (ctags.Parser, error) {
			parser, err := ctags.NewParser(ctagsCommand)
			if err != nil {
				return nil, errors.Wrap(err, fmt.Sprintf("command: %s", ctagsCommand))
			}
			return parser, nil
		},
		Path: cacheDir,
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
	handler := nethttp.Middleware(opentracing.GlobalTracer(), service.Handler())

	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	server := &http.Server{Addr: addr, Handler: handler}
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
