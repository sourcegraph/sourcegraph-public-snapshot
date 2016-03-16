package sgx

import (
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
)

func init() {
	g, err := cli.CLI.AddCommand("internal",
		"internal subcommands",
		"The internal group contains internal commands not intended for users.",
		&reposCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	cli.Internal = g

	initStressTest(g)
}
