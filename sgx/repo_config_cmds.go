package sgx

import (
	"encoding/json"
	"fmt"
	"log"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
)

func initRepoConfigCmds(repoGroup *flags.Command) {
	g, err := repoGroup.AddCommand("config",
		"manage repo config",
		"The 'src repo config' group's commands manage the configuration of repositories.",
		&repoConfigCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.AddCommand("get",
		"get a repo's config",
		"The get subcommand gets a repository's configuration.",
		&repoConfigGetCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.AddCommand("enable",
		"enable a repo",
		"The enable subcommand enables a repository.",
		&repoConfigEnableCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = g.AddCommand("disable",
		"disable a repo",
		"The disable subcommand disables a repository.",
		&repoConfigDisableCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type repoConfigCmd struct{}

func (c *repoConfigCmd) Execute(args []string) error { return nil }

type repoConfigGetCmd struct {
	Args struct {
		URI string `name:"REPO-URI" description:"repository URI (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes" count:"1"`
}

func (c *repoConfigGetCmd) Execute(args []string) error {
	cl := Client()

	repoSpec := &sourcegraph.RepoSpec{URI: c.Args.URI}

	conf, err := cl.Repos.GetConfig(cliCtx, repoSpec)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(conf, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	return nil
}

type repoConfigEnableCmd struct {
	Args struct {
		URI string `name:"REPO-URI" description:"repository URI (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes" count:"1"`
}

func (c *repoConfigEnableCmd) Execute(args []string) error {
	return repoConfigEnableOrDisable(c.Args.URI, true)
}

type repoConfigDisableCmd struct {
	Args struct {
		URI string `name:"REPO-URI" description:"repository URI (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes" count:"1"`
}

func (c *repoConfigDisableCmd) Execute(args []string) error {
	return repoConfigEnableOrDisable(c.Args.URI, false)
}

func repoConfigEnableOrDisable(repoURI string, enable bool) error {
	cl := Client()

	repoSpec := &sourcegraph.RepoSpec{URI: repoURI}

	var meth func(context.Context, *sourcegraph.RepoSpec, ...grpc.CallOption) (*pbtypes.Void, error)
	if enable {
		meth = cl.Repos.Enable
	} else {
		meth = cl.Repos.Disable
	}

	if _, err := meth(cliCtx, repoSpec); err != nil {
		return err
	}
	return nil
}
