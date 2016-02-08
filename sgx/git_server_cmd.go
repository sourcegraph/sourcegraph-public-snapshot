package sgx

import (
	"log"
	"net"
	"net/http"

	"src.sourcegraph.com/sourcegraph/pkg/gitserver"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
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
}

func (c *gitServerCmd) Execute(args []string) error {
	gitserver.RegisterHandler()

	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return err
	}
	log.Printf("Git server listening on %s", l.Addr())
	return http.Serve(l, nil)
}
