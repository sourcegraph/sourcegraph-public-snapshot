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
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

var addr = os.Getenv("SRC_ADDR")                    // RPC listen address for git server
var reposDir = os.Getenv("SRC_REPOS_DIR")           // root dir containing repos
var autoTerminate = os.Getenv("SRC_AUTO_TERMINATE") // terminate if stdin gets closed (e.g. parent process died)
var profBindAddr = os.Getenv("SRC_PROF_HTTP")       // net/http/pprof http bind address

func main() {
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

	if addr == "" {
		addr = "127.0.0.1:0"
	}
	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("git-server: listening on %s\n", l.Addr())
	log.Fatal(gitserver.Serve(l))
}
