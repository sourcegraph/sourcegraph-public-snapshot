package cli

import (
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	approuter "sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/cli/cli"
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
	// TODO(mate): add argument to select language
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
		def.Def.DefKey.CommitID = ""
		fmt.Printf("https://sourcegraph.com%s %d\n", approuter.Rel.DefKeyToLandURL(def.Def.DefKey).String(), defResult.RefCount)
	}

	return nil
}
