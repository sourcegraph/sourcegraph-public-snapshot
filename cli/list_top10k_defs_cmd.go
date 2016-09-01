package cli

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/internal/coverage"
)

func init() {
	_, err := cli.CLI.AddCommand("top10k_defs",
		"list the top10k defs along with their refcounts",
		"List the top10k Go definitions that are indexed by Sourcegraph; used for sitemap generation",
		&listTop10kDefsCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type listTop10kDefsCmd struct {
	backend coverage.Client
	cl      *sourcegraph.Client
}

func (c *listTop10kDefsCmd) Execute(args []string) error {
	results, err := cliClient.Search.Search(cliContext, &sourcegraph.SearchOp{
		Opt: &sourcegraph.SearchOptions{
			Languages:    []string{"Go"},
			NotKinds:     []string{"package"},
			IncludeRepos: false,
			ListOptions:  sourcegraph.ListOptions{PerPage: 20},
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
