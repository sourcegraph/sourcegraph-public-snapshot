package sgx

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"code.google.com/p/rog-go/parallel"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/sgx/cli"
	"sourcegraph.com/sourcegraph/sourcegraph/util/buildutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/textutil"
	"sourcegraph.com/sourcegraph/sourcegraph/util/timeutil"
)

func init() {
	reposGroup, err := cli.CLI.AddCommand("repo",
		"manage repos",
		"The repo subcommands manage repos.",
		&reposCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	reposGroup.Aliases = []string{"repos", "r"}

	_, err = reposGroup.AddCommand("get",
		"get a repo",
		"The `sgx repo get` command gets a repo.",
		&repoGetCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	listC, err := reposGroup.AddCommand("list",
		"list repos",
		"The `sgx repo list` command lists repos.",
		&repoListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	listC.Aliases = []string{"ls"}

	_, err = reposGroup.AddCommand("create",
		"create a repo",
		"The `sgx repo create` command creates a new repo.",
		&repoCreateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = reposGroup.AddCommand("link",
		"symlink an existing repo",
		"The `sgx repo link` command creates a symlink to an existing local repo.",
		&repoLinkCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	deleteC, err := reposGroup.AddCommand("delete",
		"delete a repo",
		"The `sgx repo rm` command deletes a repo.",
		&repoDeleteCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	deleteC.Aliases = []string{"rm"}

	_, err = reposGroup.AddCommand("sync",
		"syncs repos and triggers builds for recent commits",
		`The 'sgx repo sync' command syncs repos and triggers builds for recent commits.

If multiple REPO-URIs are provided, the syncs are performed concurrently.


TIPS

Sync all of a person/org's repos:

	sgx repo sync `+"`"+`sgx repo list --owner USER`+"`"+`

Same as above, but for a deployed site:

	sgx env exec sgx repo sync `+"`"+`sgx env exec sgx repo list --owner USER`+"`"+`
`,
		&repoSyncCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = reposGroup.AddCommand("refresh-vcs",
		"refresh repo VCS data",
		"The 'sgx repo refresh-vcs' command refreshes VCS data for the specified repositories.",
		&repoRefreshVCSCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	initRepoConfigCmds(reposGroup)
}

type reposCmd struct{}

func (c *reposCmd) Execute(args []string) error { return nil }

type repoGetCmd struct {
	Args struct {
		URI string `name:"REPO-URI" description:"repository URI (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes" count:"1"`

	Config bool `long:"config" description:"also get repo config"`
}

func (c *repoGetCmd) Execute(args []string) error {
	cl := Client()

	repoSpec := &sourcegraph.RepoSpec{URI: c.Args.URI}

	repo, err := cl.Repos.Get(cliCtx, repoSpec)
	if err != nil {
		return err
	}

	b, err := json.MarshalIndent(repo, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(b))

	if c.Config {
		conf, err := cl.Repos.GetConfig(cliCtx, repoSpec)
		if err != nil {
			return err
		}
		log.Println()
		log.Println("# Config")
		b, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(b))
	}

	return nil
}

type repoListCmd struct {
	Owner string `long:"owner" description:"login of user whose owned repositories to list"`
	Query string `short:"q" long:"query" description:"query"`
}

func (c *repoListCmd) Execute(args []string) error {
	cl := Client()

	for page := 1; ; page++ {
		repos, err := cl.Repos.List(cliCtx, &sourcegraph.RepoListOptions{
			Owner:       c.Owner,
			Query:       c.Query,
			ListOptions: sourcegraph.ListOptions{Page: int32(page)},
		})

		if err != nil {
			return err
		}
		if len(repos.Repos) == 0 {
			break
		}
		for _, repo := range repos.Repos {
			fmt.Println(repo.URI)
		}
	}
	return nil
}

type repoCreateCmd struct {
	Args struct {
		URI string `name:"REPO-URI" description:"desired repository URI (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes" count:"1"`
	VCS      string `long:"vcs" description:"git or hg" default:"git" required:"yes"`
	CloneURL string `short:"u" long:"clone-url" description:"clone URL of existing repo (if this repo is a mirror)"`
	Mirror   bool   `short:"m" long:"mirror" description:"create the repo as a mirror"`
}

func (c *repoCreateCmd) Execute(args []string) error {
	cl := Client()

	repo, err := cl.Repos.Create(cliCtx, &sourcegraph.ReposCreateOp{
		URI:      c.Args.URI,
		VCS:      c.VCS,
		CloneURL: c.CloneURL,
		Mirror:   c.Mirror,
	})
	if err != nil {
		return err
	}
	log.Printf("# created: %s", repo.URI)
	return nil
}

type repoLinkCmd struct {
	Args struct {
		URI string `name:"REPO-URI" description:"desired repository URI (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes" count:"1"`
	Path   string `short:"p" long:"path" description:"path to local repo" default:"."`
	SGPath string `short:"s" long:"sgpath" description:"path to Sourcegraph dir" default:"$SGPATH"`
}

func (c *repoLinkCmd) Execute(args []string) error {
	c.Args.URI = filepath.Clean(c.Args.URI)
	if absPath, err := filepath.Abs(c.Path); err == nil {
		c.Path = filepath.Clean(absPath)
	} else {
		return err
	}
	c.SGPath = os.ExpandEnv(c.SGPath)
	if c.SGPath == "" {
		return fmt.Errorf("SGPATH not set. Use `-s` flag to specify path to Sourcegraph dir (eg. `-s ~/.sourcegraph`)")
	}
	_, err := os.Stat(c.SGPath)
	if os.IsNotExist(err) {
		return fmt.Errorf("Path %s does not exist. Use `-s` flag to specify path to Sourcegraph dir (eg. `-s ~/.sourcegraph`)", c.SGPath)
	}
	repoPath := filepath.Join(c.SGPath, "repos", c.Args.URI)
	repoDir := filepath.Dir(repoPath)
	if err := os.MkdirAll(repoDir, 0755); err != nil {
		return err
	}
	if err := os.Symlink(c.Path, repoPath); err != nil {
		return err
	}
	log.Printf("# symlinked repo '%s' to local dir '%s'", c.Args.URI, c.Path)
	log.Printf("#   %s -> %s'", repoPath, c.Path)
	return nil
}

type repoDeleteCmd struct {
	Args struct {
		URIs []string `name:"REPO-URIs" description:"repository URIs to delete (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes"`
}

func (c *repoDeleteCmd) Execute(args []string) error {
	cl := Client()

	for _, uri := range c.Args.URIs {
		if _, err := cl.Repos.Delete(cliCtx, &sourcegraph.RepoSpec{URI: uri}); err != nil {
			return err
		}
		log.Printf("# deleted: %s", uri)
	}
	return nil
}

type repoSyncCmd struct {
	Args struct {
		URIs []string `name:"REPO-URI" description:"repository URIs (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes"`
}

func (c *repoSyncCmd) Execute(args []string) error {
	par := parallel.NewRun(20)
	for _, repo_ := range c.Args.URIs {
		repo := repo_
		par.Do(func() error {
			if err := c.sync(repo); err != nil {
				return fmt.Errorf(red("%s:")+" %s", repo, err)
			}
			return nil
		})
	}
	if err := par.Wait(); err != nil {
		if errs, ok := err.(parallel.Errors); ok {
			for _, err := range errs {
				log.Println(err)
			}
			return fmt.Errorf("encountered %d errors (see above)", len(errs))
		}
		return err
	}
	return nil
}

func (c *repoSyncCmd) sync(repoURI string) error {
	log := log.New(os.Stderr, cyan(strings.TrimPrefix(strings.TrimPrefix(repoURI+": ", "github.com/"), "sourcegraph.com/")), 0)

	cl := Client()

	repoSpec := sourcegraph.RepoSpec{URI: repoURI}

	repo, err := cl.Repos.Get(cliCtx, &repoSpec)
	if err != nil {
		return err
	}

	conf, err := cl.Repos.GetConfig(cliCtx, &repoSpec)
	if err != nil {
		return err
	}

	if !conf.Enabled {
		return fmt.Errorf("repo %s is not enabled", repoURI)
	}

	repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: repo.DefaultBranch}
	commit, err := cl.Repos.GetCommit(cliCtx, &repoRevSpec)
	if err != nil {
		return err
	}
	repoRevSpec.CommitID = string(commit.ID)
	log.Printf("Got latest commit %s (%s): %s (%s %s).", commit.ID[:8], repo.DefaultBranch, textutil.Truncate(50, commit.Message), commit.Author.Email, timeutil.TimeAgo(commit.Author.Date))

	buildInfo, err := cl.Builds.GetRepoBuildInfo(cliCtx, &sourcegraph.BuildsGetRepoBuildInfoOp{Repo: repoRevSpec, Opt: nil})
	if err != nil {
		return err
	}
	if buildInfo != nil && buildInfo.LastSuccessfulCommit != nil && buildInfo.LastSuccessfulCommit.ID == commit.ID {
		log.Printf("Latest commit is already built.")
	} else {
		if buildInfo == nil || buildInfo.LastSuccessfulCommit == nil {
			log.Printf("No builds found.")
		} else if buildInfo.LastSuccessfulCommit != nil {
			log.Printf("Most recent build was for commit %s (%s).", buildInfo.LastSuccessfulCommit.ID[:8], timeutil.TimeAgo(buildInfo.LastSuccessfulCommit.Author.Date))
		}
		b, err := cl.Builds.Create(cliCtx, &sourcegraph.BuildsCreateOp{RepoRev: repoRevSpec, Opt: &sourcegraph.BuildCreateOptions{
			BuildConfig: sourcegraph.BuildConfig{
				Import:   true,
				Queue:    true,
				Priority: int32(buildutil.DefaultPriority(repo.Private, 0) + 10),
			},
		}})

		if err != nil {
			return err
		}
		log.Printf("Created build #%s for commit %s.", b.Spec().IDString(), commit.ID[:8])
	}

	return nil
}

type repoRefreshVCSCmd struct {
	Args struct {
		URIs []string `name:"REPO-URI" description:"repository URIs (e.g., host.com/myrepo)"`
	} `positional-args:"yes" required:"yes"`
}

func (c *repoRefreshVCSCmd) Execute(args []string) error {
	cl := Client()
	for _, repoURI := range c.Args.URIs {
		repo, err := cl.Repos.Get(cliCtx, &sourcegraph.RepoSpec{URI: repoURI})
		if err != nil {
			return err
		}

		repoRevSpec := sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: repo.DefaultBranch}
		preCommit, err := cl.Repos.GetCommit(cliCtx, &repoRevSpec)
		if err != nil {
			return err
		}

		if _, err := cl.MirrorRepos.RefreshVCS(cliCtx, &sourcegraph.MirrorReposRefreshVCSOp{Repo: repoRevSpec.RepoSpec}); err != nil {
			return err
		}

		postCommit, err := cl.Repos.GetCommit(cliCtx, &repoRevSpec)
		if err != nil {
			return err
		}

		if preCommit.ID == postCommit.ID {
			log.Printf("%s: latest commit on %s unchanged: %s %s", repo.URI, repo.DefaultBranch, preCommit.ID, timeutil.TimeAgo(preCommit.Author.Date))
		} else {
			log.Printf("%s: updated latest commit on %s: %s %s (was %s %s)", repo.URI, repo.DefaultBranch, postCommit.ID, timeutil.TimeAgo(postCommit.Author.Date), preCommit.ID, timeutil.TimeAgo(preCommit.Author.Date))
		}
	}
	return nil
}
