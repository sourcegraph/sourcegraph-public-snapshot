// gitserver is the gitserver server.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver"

//docker:install git openssh-client

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"os"
	"os/signal"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

var reposDir = env.Get("SRC_REPOS_DIR", "", "Root dir containing repos.")
var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")

func main() {
	env.Lock()

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGHUP)
		<-c
		os.Exit(0)
	}()

	if reposDir == "" {
		log.Fatal("git-server: SRC_REPOS_DIR is required")
	}
	gitserver := gitserver.Server{ReposDir: reposDir}

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log.Printf("Profiler available on %s/pprof", profBindAddr)
	}

	l, err := net.Listen("tcp", ":3178")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("git-server: listening on %s\n", l.Addr())
	log.Fatal(gitserver.Serve(l))
}
