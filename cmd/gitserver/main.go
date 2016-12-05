// gitserver is the gitserver server.
package main // import "sourcegraph.com/sourcegraph/sourcegraph/cmd/gitserver"

//docker:install git openssh-client

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/env"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

var addr = env.Get("SRC_ADDR", "127.0.0.1:0", "RPC listen address for git server.")
var reposDir = env.Get("SRC_REPOS_DIR", "", "Root dir containing repos.")
var autoTerminate = env.Get("SRC_AUTO_TERMINATE", "false", "Terminate if stdin gets closed (e.g. parent process died).")
var profBindAddr = env.Get("SRC_PROF_HTTP", "", "net/http/pprof http bind address.")

func main() {
	env.Lock()

	if reposDir == "" {
		log.Fatal("git-server: SRC_REPOS_DIR is required")
	}
	gitserver := gitserver.Server{ReposDir: reposDir}

	if b, _ := strconv.ParseBool(autoTerminate); b {
		go func() {
			io.Copy(ioutil.Discard, os.Stdin)
			log.Fatal("git-server: stdin closed, terminating")
		}()
	}

	if profBindAddr != "" {
		go debugserver.Start(profBindAddr)
		log.Printf("Profiler available on %s/pprof", profBindAddr)
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("git-server: listening on %s\n", l.Addr())
	log.Fatal(gitserver.Serve(l))
}
