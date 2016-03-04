package sgx

import (
	"fmt"
	"log"

	"src.sourcegraph.com/sourcegraph/sgx/cli"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/fed/discover"
)

func init() {
	g, err := cli.CLI.AddCommand("discover",
		"discovery",
		"The 'sgx discover' commands perform discovery on repos, sites, etc.",
		&discoverCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.AddCommand("repo",
		"perform repo discovery",
		"The 'sgx discover repo' command performs the repo discovery process to determine the Sourcegraph instance that hosts a specified repository.",
		&discoverRepoCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type discoverCmd struct{}

func (c *discoverCmd) Execute(args []string) error { return nil }

type discoverRepoCmd struct {
	Args struct {
		Paths []string `name:"PATH" description:"paths to repositories or other resources (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes"`
}

func (c *discoverRepoCmd) Execute(args []string) error {
	if len(c.Args.Paths) == 0 {
		log.Println("# warning: nothing to discover")
	}
	for _, repo := range c.Args.Paths {
		fmt.Print(repo, ": ")
		info, err := discover.Repo(context.Background(), repo)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}
		fmt.Println(info)
	}
	return nil
}
