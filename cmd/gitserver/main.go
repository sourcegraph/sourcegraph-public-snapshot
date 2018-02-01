//docker:install git openssh-client
//docker:user sourcegraph /data/repos

// gitserver is the gitserver server.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver"

import (
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"os"
	"os/signal"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/tracer"
)

const repoCleanupInterval = 24 * time.Hour

var (
	reposDir          = env.Get("SRC_REPOS_DIR", "", "Root dir containing repos.")
	profBindAddr      = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")
	runRepoCleanup, _ = strconv.ParseBool(env.Get("SRC_RUN_REPO_CLEANUP", "", "Periodically remove inactive repositories."))
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init("gitserver")

	// Filter log output by level.
	lvl, err := log15.LvlFromString(env.LogLevel)
	if err == nil {
		log15.Root().SetHandler(log15.LvlFilterHandler(lvl, log15.StderrHandler))
	}

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c
		os.Exit(0)
	}()

	if reposDir == "" {
		log.Fatal("git-server: SRC_REPOS_DIR is required")
	}
	gitserver := server.Server{
		ReposDir:            reposDir,
		MaxConcurrentClones: conf.Get().GitMaxConcurrentClones,
	}
	gitserver.RegisterMetrics()

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log.Printf("Profiler available on %s/pprof", profBindAddr)
	}

	if runRepoCleanup {
		go func() {
			for {
				gitserver.CleanupRepos()
				time.Sleep(repoCleanupInterval)
			}
		}()
	}

	handler := nethttp.Middleware(opentracing.GlobalTracer(), gitserver.Handler())

	log15.Info("git-server: listening", "addr", ":3178")
	srv := &http.Server{Addr: ":3178", Handler: handler}
	log.Fatal(srv.ListenAndServe())
}
