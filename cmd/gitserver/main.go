// gitserver is the gitserver server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/gitserver"

import (
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/opentracing-contrib/go-stdlib/nethttp"
	"github.com/opentracing/opentracing-go"
	"gopkg.in/inconshreveable/log15.v2"

	"github.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"github.com/sourcegraph/sourcegraph/pkg/debugserver"
	"github.com/sourcegraph/sourcegraph/pkg/env"
	"github.com/sourcegraph/sourcegraph/pkg/tracer"
)

var (
	reposDir          = env.Get("SRC_REPOS_DIR", "/data/repos", "Root dir containing repos.")
	runRepoCleanup, _ = strconv.ParseBool(env.Get("SRC_RUN_REPO_CLEANUP", "", "Periodically remove inactive repositories."))
	mountPoint        = env.Get("SRC_REPOS_MOUNT_POINT", "/data/repos", "Where the disk containing $SRC_REPOS_DIR is mounted")
	wantFreeG         = env.Get("SRC_REPOS_DESIRED_FREE_GB", "10", "How many gigabytes of space to keep free on the disk with the repos")
	janitorInterval   = env.Get("SRC_REPOS_JANITOR_INTERVAL", "1m", "Interval between cleanup runs")
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

	if !isSubPath(mountPoint, reposDir) {
		log.Fatalf("$SRC_REPOS_DIR is %s, want a subdirectory of $SRC_REPOS_MOUNT_POINT (%s)", reposDir, mountPoint)
	}

	wantFreeG2, err := strconv.Atoi(wantFreeG)
	if err != nil {
		log.Fatalf("parsing $SRC_REPOS_DESIRED_FREE_GB: %v", err)
	}
	gitserver := server.Server{
		ReposDir:                reposDir,
		DeleteStaleRepositories: runRepoCleanup,
		MountPoint:              mountPoint,
		DesiredFreeDiskSpace:    uint64(wantFreeG2 * 1024 * 1024 * 1024),
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

	janitorInterval2, err := time.ParseDuration(janitorInterval)
	if err != nil {
		log.Fatalf("parsing $SRC_REPOS_JANITOR_INTERVAL: %v", err)
	}
	go func() {
		for {
			gitserver.Janitor()
			time.Sleep(janitorInterval2)
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

// isSubPath returns true if dir could contain findme, based only on lexical
// properties of the paths.
func isSubPath(dir, findme string) bool {
	rel, err := filepath.Rel(dir, findme)
	return err == nil && !strings.HasPrefix(rel, "..")
}
