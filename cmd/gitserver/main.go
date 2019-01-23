// gitserver is the gitserver server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/gitserver"

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	opentracing "github.com/opentracing/opentracing-go"
	log15 "gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

const janitorInterval = 24 * time.Hour

var (
	reposDir          = env.Get("SRC_REPOS_DIR", "/data/repos", "Root dir containing repos.")
	runRepoCleanup, _ = strconv.ParseBool(env.Get("SRC_RUN_REPO_CLEANUP", "", "Periodically remove inactive repositories."))
)

func main() {
	env.Lock()
	env.HandleHelpFlag()
	tracer.Init()

	if reposDir == "" {
		log.Fatal("git-server: SRC_REPOS_DIR is required")
	}
	if err := os.MkdirAll(reposDir, os.ModePerm); err != nil {
		log.Fatalf("failed to create SRC_REPOS_DIR: %s", err)
	}

	gitserver := server.Server{
		ReposDir:                reposDir,
		DeleteStaleRepositories: runRepoCleanup,
	}
	gitserver.RegisterMetrics()

	if tmpDir, err := gitserver.SetupAndClearTmp(); err != nil {
		log.Fatalf("failed to setup temporary directory: %s", err)
	} else {
		// Additionally set TMP_DIR so other temporary files we may accidently
		// create are on the faster RepoDir mount.
		os.Setenv("TMP_DIR", tmpDir)
	}

	// Create Handler now since it also initializes state
	handler := nethttp.Middleware(opentracing.GlobalTracer(), gitserver.Handler())

	go debugserver.Start()

	go func() {
		for {
			gitserver.Janitor()
			time.Sleep(janitorInterval)
		}
	}()

	port := "3178"
	host := ""
	if env.InsecureDev {
		host = "127.0.0.1"
	}
	addr := net.JoinHostPort(host, port)
	srv := &http.Server{Addr: addr, Handler: handler}
	log15.Info("git-server: listening", "addr", srv.Addr)

	go func() {
		err := srv.ListenAndServe()
		if err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	// Listen for shutdown signals. When we receive one attempt to clean up,
	// but do an insta-shutdown if we receive more than one signal.
	c := make(chan os.Signal, 2)
	signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
	<-c
	go func() {
		<-c
		os.Exit(0)
	}()

	// Stop accepting requests. In the future we should use graceful shutdown.
	srv.Close()

	// The most important thing this does is kill all our clones. If we just
	// shutdown they will be orphaned and continue running.
	gitserver.Stop()
}
