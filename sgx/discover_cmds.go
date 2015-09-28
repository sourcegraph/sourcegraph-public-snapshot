package sgx

import (
	"fmt"
	"log"

	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/fed/discover"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
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

	_, err = g.AddCommand("site",
		"perform site discovery",
		"The 'sgx discover site' command performs the site discovery process to determine the Sourcegraph server configuration (endpoint URLs, etc.) for the specified host (and optional port).",
		&discoverSiteCmd{},
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

type discoverSiteCmd struct {
	Args struct {
		Hosts []string `name:"HOST" description:"Sourcegraph hosts (e.g., host.com, host.com:1234)"`
	} `positional-args:"yes" required:"yes"`
}

func (c *discoverSiteCmd) Execute(args []string) error {
	if len(c.Args.Hosts) == 0 {
		log.Println("# warning: nothing to discover")
	}
	for _, host := range c.Args.Hosts {
		fmt.Print(host, ": ")
		info, err := discover.Site(context.Background(), host)
		if err != nil {
			fmt.Printf("error: %s\n", err)
			continue
		}
		fmt.Println(info)
	}
	return nil
}
