package sgx

import (
	"fmt"
	"log"

	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
)

func init() {
	_, err := cli.CLI.AddCommand("q",
		"perform a search",
		"The 'src q' subcommand searches the Sourcegraph server for a given query. It returns code, repo, user, etc., results.",
		&queryCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type queryCmd struct {
	Args struct {
		Query []string `name:"query" description:"search query"`
	} `positional-args:"yes"`
}

func (c *queryCmd) Execute(args []string) error {
	cl := Client()

	query := strings.Join(c.Args.Query, " ")
	res, err := cl.Search.Search(cliCtx, &sourcegraph.SearchOptions{
		Query:  query,
		Defs:   true,
		Repos:  true,
		People: true,
	})
	if err != nil {
		return err
	}

	if len(res.Defs) > 0 {
		fmt.Println("# Defs")
		for _, def := range res.Defs {
			kw := def.FmtStrings.DefKeyword
			if kw != "" {
				kw += " "
			}
			fmt.Println(kw+bold(yellow(def.FmtStrings.Name.ScopeQualified))+def.FmtStrings.NameAndTypeSeparator+strings.TrimSpace(def.FmtStrings.Type.ScopeQualified), "\t", fade(def.Repo))
		}
		fmt.Println()
	}

	if len(res.Repos) > 0 {
		fmt.Println("# Repositories")
		for _, repo := range res.Repos {
			fmt.Println(repo.URI)
		}
		fmt.Println()
	}

	if len(res.People) > 0 {
		fmt.Println("# People")
		for _, p := range res.People {
			fmt.Println(p.Login)
		}
		fmt.Println()
	}

	if len(res.Defs) == 0 && len(res.Repos) == 0 && len(res.People) == 0 {
		log.Printf("# No results found for %q", query)
	}

	return nil
}
