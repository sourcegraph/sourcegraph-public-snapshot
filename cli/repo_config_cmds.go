package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"sourcegraph.com/sourcegraph/go-flags"
	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
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

	_, err = g.AddCommand("app",
		"configure a repo app",
		"The app subcommand configures a repo app (enabling or disabling it).",
		&repoConfigAppCmd{},
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
	cl := cliClient

	repoSpec := &sourcegraph.RepoSpec{URI: c.Args.URI}

	conf, err := cl.Repos.GetConfig(cliContext, repoSpec)
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

type repoConfigAppCmd struct {
	Args struct {
		Repo string `name:"REPO" description:"repository URI (e.g., host.com/myrepo)"`
		App  string `name:"APP" description:"application ID"`
	} `positional-args:"yes" required:"yes" count:"1"`

	Enable  bool `long:"enable" description:"enable app"`
	Disable bool `long:"disable" description:"disable app"`
}

func (c *repoConfigAppCmd) Execute(args []string) error {
	cl := cliClient

	if (!c.Enable && !c.Disable) || (c.Enable && c.Disable) {
		return errors.New("exactly one of --enable and --disable must be specified")
	}

	repo, err := cl.Repos.Get(cliContext, &sourcegraph.RepoSpec{URI: c.Args.Repo})
	if err != nil {
		return err
	}

	_, err = cl.Repos.ConfigureApp(cliContext, &sourcegraph.RepoConfigureAppOp{
		Repo:   repo.URI,
		App:    c.Args.App,
		Enable: c.Enable,
	})
	if err != nil {
		return err
	}

	var verb string
	switch {
	case c.Enable:
		verb = "enabled"
	case c.Disable:
		verb = "disabled"
	}
	fmt.Printf("# %s app %s for %s\n", verb, c.Args.App, repo.URI)

	return nil
}
