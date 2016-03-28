package sgx

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserverlegacy"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
)

func init() {
	_, err := cli.CLI.AddCommand("git-server",
		"run git server",
		"A server to run git commands remotely.",
		&gitServerCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type gitServerCmd struct {
	Addr          string `long:"addr" default:"" description:"RPC listen address for git server" env:"SRC_ADDR"`
	NewAddr       string `long:"new-addr" default:"127.0.0.1:0" description:"RPC listen address for new git server" env:"SRC_NEW_ADDR"`
	ReposDir      string `long:"repos-dir" description:"root dir containing repos" env:"SRC_REPOS_DIR"`
	AutoTerminate bool   `long:"auto-terminate" description:"terminate if stdin gets closed (e.g. parent process died)" env:"SRC_AUTO_TERMINATE"`
	ProfBindAddr  string `long:"prof-http" description:"net/http/pprof http bind address" value-name:"BIND-ADDR" env:"SRC_PROF_HTTP"`
}

func (c *gitServerCmd) Execute(args []string) error {
	if c.ReposDir == "" {
		log.Fatal("git-server: --repos-dir flag is required")
	}
	gitserver.ReposDir = c.ReposDir
	gitserverlegacy.ReposDir = c.ReposDir

	if c.AutoTerminate {
		go func() {
			io.Copy(ioutil.Discard, os.Stdin)
			log.Fatal("git-server: stdin closed, terminating")
		}()
	}

	if c.ProfBindAddr != "" {
		startDebugServer(c.ProfBindAddr)
	}

	if c.Addr != "" {
		go func() {
			gitserverlegacy.RegisterHandler()

			l, err := net.Listen("tcp", c.Addr)
			if err != nil {
				log.Print(err)
				return
			}
			log.Print(http.Serve(l, nil))
		}()
	}

	if c.NewAddr != "" {
		go func() {
			l, err := net.Listen("tcp", c.NewAddr)
			if err != nil {
				log.Fatal(err)
				return
			}
			fmt.Printf("git-server: listening on %s\n", l.Addr())
			log.Fatal(gitserver.Serve(l))
		}()
	}

	select {}
}
