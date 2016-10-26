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

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/debugserver"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/gitserver"
)

func run() error {
	log.SetFlags(0)
	log.SetPrefix("")

	cli := flags.NewNamedParser("gitserver", flags.PrintErrors|flags.PassDoubleDash)

	_, err := cli.AddCommand("run",
		"run git server",
		"A server to run git commands remotely.",
		&runCmd{},
	)
	if err != nil {
		return err
	}

	_, err = cli.Parse()
	return err
}

func main() {
	err := run()
	if err != nil {
		log.Fatalln(err)
	}
}

type runCmd struct {
	Addr          string `long:"addr" default:"127.0.0.1:0" description:"RPC listen address for git server" env:"SRC_ADDR"`
	ReposDir      string `long:"repos-dir" description:"root dir containing repos" env:"SRC_REPOS_DIR"`
	AutoTerminate bool   `long:"auto-terminate" description:"terminate if stdin gets closed (e.g. parent process died)" env:"SRC_AUTO_TERMINATE"`
	ProfBindAddr  string `long:"prof-http" description:"net/http/pprof http bind address" value-name:"BIND-ADDR" env:"SRC_PROF_HTTP"`
}

func (c *runCmd) Execute(args []string) error {
	if c.ReposDir == "" {
		log.Fatalln("gitserver: --repos-dir flag is required")
	}
	gitserver := gitserver.Server{ReposDir: c.ReposDir}

	if c.AutoTerminate {
		go func() {
			io.Copy(ioutil.Discard, os.Stdin)
			log.Fatalln("gitserver: stdin closed, terminating")
		}()
	}

	if c.ProfBindAddr != "" {
		go debugserver.Start(c.ProfBindAddr)
		log.Printf("gitserver: profiler available on %s/pprof", c.ProfBindAddr)
	}

	if c.Addr != "" {
		go func() {
			l, err := net.Listen("tcp", c.Addr)
			if err != nil {
				log.Fatalln(err)
			}
			fmt.Printf("gitserver: listening on %s\n", l.Addr())
			log.Fatalln(gitserver.Serve(l))
		}()
	}

	select {}
}
