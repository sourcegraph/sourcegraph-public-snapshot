// gitserver is the gitserver server.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver"

//docker:install git openssh-client

import (
	"log"
	"net/http"
	"strconv"
	"syscall"
	"time"

	"os"
	"os/signal"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

const repoCleanupInterval = 24 * time.Hour

var reposDir = env.Get("SRC_REPOS_DIR", "", "Root dir containing repos.")
var githubPersonalAccessToken = env.Get("GITHUB_PERSONAL_ACCESS_TOKEN", "", "personal access token for GitHub. All requests will use this token to access the Github API. Used give access to private Github code on a private server.")
var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")
var runRepoCleanup, _ = strconv.ParseBool(env.Get("SRC_RUN_REPO_CLEANUP", "", "Periodically remove inactive repositories."))

func main() {
	env.Lock()
	env.HandleHelpFlag()

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
		ReposDir:          reposDir,
		GithubAccessToken: githubPersonalAccessToken,
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

	log.Print("git-server: listening on :3178")
	srv := &http.Server{Addr: ":3178", Handler: gitserver.Handler()}
	log.Fatal(srv.ListenAndServe())
}
