package sgx

import (
	"log"

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
	return gitserver.ListenAndServe()
}
