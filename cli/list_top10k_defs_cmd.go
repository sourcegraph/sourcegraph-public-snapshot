package cli

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/coverage"
)

func init() {
	_, err := cli.CLI.AddCommand("list_top_defs",
		"list the top defs along with their refcounts",
		"List the top Go definitions that are indexed by Sourcegraph. Used for sitemap generation.",
		&listTopDefsCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type listTopDefsCmd struct {
	Limit int `long:"limit" description:"max number of defs to list" default:"100"`

	backend coverage.Client
	cl      *sourcegraph.Client
}

func (c *listTopDefsCmd) Execute(args []string) error {
	results, err := cliClient.Search.Search(cliContext, &sourcegraph.SearchOp{
		Opt: &sourcegraph.SearchOptions{
			Languages:    []string{"Go"},
			NotKinds:     []string{"package"},
			IncludeRepos: false,
			ListOptions:  sourcegraph.ListOptions{PerPage: int32(c.Limit)},
			AllowEmpty:   true,
		},
	})
	if err != nil {
		return err
	}

	for _, defResult := range results.DefResults {
		def := &defResult.Def
		fmt.Printf("%s %d\n", def.Def.Name, defResult.RefCount)
	}

	return nil
}
