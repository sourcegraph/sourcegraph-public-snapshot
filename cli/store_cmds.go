package cli

import (
	"fmt"
	"log"
	"os"
	"path"

	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/graphstoreutil"
	srclib "sourcegraph.com/sourcegraph/srclib/cli"
)

func init() {
	storeCmdInst := &storeCmd{}
	storeC, err := cli.CLI.AddCommand("store",
		"server-side graph store commands",
		`The store subcommands import, manage, index, and query server-side graph store data. The client interface is available in the subcommands of 'srclib store'.`,
		storeCmdInst,
	)
	if err != nil {
		log.Fatal(err)
	}

	srclib.OpenStore = func() (interface{}, error) {
		return graphstoreutil.New(os.ExpandEnv(storeCmdInst.GraphStore), nil), nil
	}

	srclib.InitStoreCmds(storeC)

	_, err = storeC.AddCommand("repo-path",
		"display the repo path for a repo",
		"The repo-path subcommand displays the repo path for a repo.",
		&storeRepoPathCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type storeCmd struct {
	GraphStore string `long:"graph-store" description:"dir in which to store graph data (defs/refs/etc.)" default:"$SGPATH/repos"`
}

func (c *storeCmd) Execute(args []string) error { return nil }

type storeRepoPathCmd struct {
	Args struct {
		URIs []string `name:"REPO-URI" description:"repository URIs (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes"`
}

func (c *storeRepoPathCmd) Execute(args []string) error {
	e := graphstoreutil.EvenlyDistributedRepoPaths{}
	for _, repoURI := range c.Args.URIs {
		fmt.Printf("%s: %s\n", repoURI, path.Join(e.RepoToPath(repoURI)...))
	}
	return nil
}
