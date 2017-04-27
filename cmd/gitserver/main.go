// gitserver is the gitserver server.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver"

//docker:install git openssh-client

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"syscall"

	"os"
	"os/signal"

	"sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver/server"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
)

var reposDir = env.Get("SRC_REPOS_DIR", "", "Root dir containing repos.")
var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")

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
	gitserver := server.Server{ReposDir: reposDir}

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log.Printf("Profiler available on %s/pprof", profBindAddr)
	}

	go func() {
		l, err := net.Listen("tcp", ":3178")
		if err != nil {
			log.Fatal(err)
		}
		log.Fatal(gitserver.ServeLegacy(l))
	}()

	fmt.Println("git-server: listening on :3278 and :3178 (legacy)")
	srv := &http.Server{Addr: ":3278", Handler: gitserver.Handler()}
	log.Fatal(srv.ListenAndServe())
}
