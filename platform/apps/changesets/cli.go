package changesets

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"

	"src.sourcegraph.com/sourcegraph/platform"
	"src.sourcegraph.com/sourcegraph/platform/putil"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/tempedit"
	"src.sourcegraph.com/sourcegraph/util/timeutil"
)

type serveFlags struct {
	ReviewGuidelines string `long:"changesets.review-guidelines" description:"loads the given file as review guidelines and displays it on the changesets page (Markdown supported)."`
	JiraURL          string `long:"jira.url" description:"URL that hosts a JIRA instance."`
	JiraCredentials  string `long:"jira.credentials" description:"HTTP basic auth credentials in the form \"user:password\" for the specified JIRA instance."`
}

var flags serveFlags

func init() {
	cli.PostInit = append(cli.PostInit, func() {
		_, err := cli.Serve.AddGroup("Changesets", "Changesets", &flags)
		if err != nil {
			log.Fatal(err)
		}
	})

	changesetsGroup, err := platform.CLI.AddCommand("changeset",
		"manage changesets",
		"The changeset subcommands manage changesets.",
		&changesetsCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	changesetsGroup.Aliases = []string{"changesets", "cs"}

	listC, err := changesetsGroup.AddCommand("list",
		"list changesets",
		"The `sgx changeset list` command lists changesets.",
		&changesetListCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
	listC.Aliases = []string{"ls"}

	_, err = changesetsGroup.AddCommand("create",
		"create a changeset",
		"The `sgx changeset create` command creates a new changeset.",
		&changesetCreateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = changesetsGroup.AddCommand("update",
		"update a changeset",
		"The `sgx changeset update` command updates a changeset.",
		&changesetUpdateCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = changesetsGroup.AddCommand("merge",
		"merge a changeset",
		"The `sgx changeset merge` command merges a changeset into its base branch on the remote.",
		&changesetMergeCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = changesetsGroup.AddCommand("close",
		"close a changeset",
		"The `sgx changeset close` command closes a changeset.",
		&changesetCloseCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}

	_, err = changesetsGroup.AddCommand("open",
		"open a changeset",
		"The `sgx changeset open` command opens a changeset.",
		&changesetOpenCmd{},
	)
	if err != nil {
		log.Fatal(err)
	}
}

type changesetsCmd struct{}

func (c *changesetsCmd) Execute(args []string) error { return nil }

type changesetsCmdCommon struct {
	// Use the Repo function instead. This ensures the repo exists
	RepoDONOTUSE string `short:"r" long:"repo" description:"repository URI"`
}

func (c *changesetsCmdCommon) Repo() (*sourcegraph.Repo, error) {
	var isGuess bool
	repoURI := c.RepoDONOTUSE
	if repoURI == "" {
		isGuess = true
		repoURI, _ = guessRepo()
		if repoURI == "" {
			return nil, errors.New("Could not guess repo, please specify a repo with -r.")
		}
	}

	cliCtx := putil.CLIContext()
	sg := sourcegraph.NewClientFromContext(cliCtx)
	repo, err := sg.Repos.Get(cliCtx, &sourcegraph.RepoSpec{URI: repoURI})
	if err != nil && isGuess && grpc.Code(err) == codes.NotFound {
		return nil, errors.New("Could not guess repo, please specify a repo with -r.")
	}
	return repo, err
}

type changesetListCmd struct {
	changesetsCmdCommon
	Status string `long:"status" description:"filter to only 'open' or 'closed' changesets (default: open)"`
}

func (c *changesetListCmd) Execute(args []string) error {
	cliCtx := putil.CLIContext()
	sg := sourcegraph.NewClientFromContext(cliCtx)

	repo, err := c.Repo()
	if err != nil {
		return err
	}

	var open, closed bool
	switch c.Status {
	case "open", "":
		open = true
	case "closed":
		closed = true
	case "all":
		open, closed = true, true
	default:
		return fmt.Errorf("Unrecognized status filter %v. Please pick one of open, closed or all", c.Status)
	}

	for page := 1; ; page++ {
		changesets, err := sg.Changesets.List(cliCtx, &sourcegraph.ChangesetListOp{
			Repo:        repo.URI,
			Open:        open,
			Closed:      closed,
			ListOptions: sourcegraph.ListOptions{Page: int32(page)},
		})

		if err != nil {
			return err
		}
		if len(changesets.Changesets) == 0 {
			if page == 1 {
				fmt.Println("no changesets for", repo.URI)
			}
			break
		}
		for _, changeset := range changesets.Changesets {
			var status string
			if changeset.ClosedAt == nil {
				status = "open"
			} else {
				status = "closed"
			}
			fmt.Printf("#%d\t%s\t@%- 10s\t%s...%s\t%s (created %s)\n", changeset.ID, status, changeset.Author.Login, changeset.DeltaSpec.Base.Rev, changeset.DeltaSpec.Head.Rev, changeset.Title, timeutil.TimeAgo(changeset.CreatedAt.Time()))
		}
	}
	return nil
}

type changesetCreateCmd struct {
	changesetsCmdCommon
	Base  string `long:"base" description:"base branch"`
	Head  string `long:"head" description:"head branch"`
	Title string `short:"t" long:"title" description:"title"`
}

func (c *changesetCreateCmd) Execute(args []string) error {
	cliCtx := putil.CLIContext()
	sg := sourcegraph.NewClientFromContext(cliCtx)

	conf, err := sg.Meta.Config(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	repo, err := c.Repo()
	if err != nil {
		return err
	}

	if c.Base == "" {
		c.Base = repo.DefaultBranch
	}
	if c.Base == "" {
		return errors.New("must specify --base (could not determine default branch for repo)")
	}
	if _, err := sg.Repos.GetCommit(cliCtx, &sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Base}); err != nil {
		return fmt.Errorf("checking that base branch %q exists on remote server: %s", c.Base, err)
	}

	if c.Head == "" {
		currentBranch, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		c.Head = strings.TrimSpace(string(currentBranch))
	}
	if c.Head == "" {
		return errors.New("must specify --head (could not determine current branch)")
	}
	remoteHeadCommit, err := sg.Repos.GetCommit(cliCtx, &sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Head})
	if err != nil {
		return fmt.Errorf("checking that head branch %q exists on remote server: %s (did you git push?)", c.Head, err)
	}

	// Convenience check that local head commit == remote head
	// commit. There are a lot of edge cases, so don't make this check
	// required.
	localHeadCommit, err := exec.Command("git", "rev-parse", c.Head).Output()
	if err != nil {
		log.Printf("warning: failed to check local head commit of branch %q: %s", c.Head, err)
	} else if lhc := strings.TrimSpace(string(localHeadCommit)); lhc != string(remoteHeadCommit.ID) {
		log.Printf("warning: local branch %q head commit is %s, remote is %s", c.Head, lhc, remoteHeadCommit.ID)
	}

	// TODO(sqs): Move this author field logic to the server so the
	// client doesn't have to fill in all of these fields.

	authInfo, err := sg.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	title, description, err := newChangesetInEditor(c.Title)
	if err != nil {
		return err
	}

	changeset, err := sg.Changesets.Create(cliCtx, &sourcegraph.ChangesetCreateOp{
		Repo: repo.RepoSpec(),
		Changeset: &sourcegraph.Changeset{
			Title:       title,
			Description: description,
			Author: sourcegraph.UserSpec{
				UID:    authInfo.UID,
				Login:  authInfo.Login,
				Domain: authInfo.Domain,
			},
			DeltaSpec: &sourcegraph.DeltaSpec{
				Base: sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Base},
				Head: sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Head},
			},
		},
	})
	if err != nil {
		return err
	}

	baseURL, err := url.Parse(conf.AppURL)
	if err != nil {
		return err
	}
	relURL, err := urlToRepoChangeset(repo.URI, changeset.ID)
	if err != nil {
		return err
	}
	log.Println(baseURL.ResolveReference(&url.URL{Path: relURL.Path[1:]}))
	return nil
}

func newChangesetInEditor(origTitle string) (title, description string, err error) {
	contents := origTitle + `
# Please enter the changeset title (in the first line) and description
# (in the subsequent lines). Lines starting with '#' will be ignored,
# and an empty message aborts the changeset.
`

	txt, err := tempedit.Edit([]byte(contents))
	if err != nil {
		return "", "", err
	}

	lines := bytes.Split(txt, []byte("\n"))
	hasTitle := false
	for _, line := range lines {
		if bytes.HasPrefix(line, []byte("#")) {
			continue
		}
		if !hasTitle {
			title = string(bytes.TrimSpace(line))
			hasTitle = true
			continue
		}
		description += string(line) + "\n"
	}
	description = strings.TrimSpace(description)

	if title == "" {
		return "", "", errors.New("aborting changeset due to empty title")
	}

	return
}

type changesetUpdateCmdCommon struct {
	changesetsCmdCommon
	Args struct {
		ID int64 `name:"ID" description:"changeset ID"`
	} `positional-args:"yes" required:"yes" count:"1"`
}

type changesetUpdateCmd struct {
	changesetUpdateCmdCommon
	Title string `short:"t" long:"title" description:"new changeset title" required:"yes"`
}

func (c *changesetUpdateCmd) Execute(args []string) error {
	cliCtx := putil.CLIContext()
	sg := sourcegraph.NewClientFromContext(cliCtx)

	repo, err := c.Repo()
	if err != nil {
		return err
	}

	ev, err := sg.Changesets.Update(cliCtx, &sourcegraph.ChangesetUpdateOp{
		Repo:  repo.RepoSpec(),
		ID:    c.Args.ID,
		Title: c.Title,
	})
	if err != nil {
		return err
	}

	log.Printf("# updated changeset %s #%d", repo.URI, ev.After.ID)
	return nil
}

type changesetMergeCmd struct {
	changesetsCmdCommon
	Squash bool `long:"squash" description:"squash multiple commits on head into a single merge commit"`
	Args   struct {
		ID int64 `name:"ID" description:"changeset ID"`
	} `positional-args:"yes" required:"yes" count:"1"`
}

func (c *changesetMergeCmd) Execute(args []string) error {
	cliCtx := putil.CLIContext()
	sg := sourcegraph.NewClientFromContext(cliCtx)

	repo, err := c.Repo()
	if err != nil {
		return err
	}

	_, err = sg.Changesets.Merge(cliCtx, &sourcegraph.ChangesetMergeOp{
		Repo:   repo.RepoSpec(),
		ID:     c.Args.ID,
		Squash: c.Squash,
	})
	if err != nil {
		return errors.New(grpc.ErrorDesc(err))
	}

	log.Printf("# merged changeset %s #%d", repo.URI, c.Args.ID)
	return nil
}

type changesetCloseCmd struct{ changesetUpdateCmdCommon }

func (c *changesetCloseCmd) Execute(args []string) error {
	cliCtx := putil.CLIContext()
	sg := sourcegraph.NewClientFromContext(cliCtx)

	repo, err := c.Repo()
	if err != nil {
		return err
	}

	ev, err := sg.Changesets.Update(cliCtx, &sourcegraph.ChangesetUpdateOp{
		Repo:  repo.RepoSpec(),
		ID:    c.Args.ID,
		Close: true,
	})
	if err != nil {
		return err
	}

	log.Printf("# closed changeset %s #%d", repo.URI, ev.After.ID)
	return nil
}

type changesetOpenCmd struct{ changesetUpdateCmdCommon }

func (c *changesetOpenCmd) Execute(args []string) error {
	cliCtx := putil.CLIContext()
	sg := sourcegraph.NewClientFromContext(cliCtx)

	repo, err := c.Repo()
	if err != nil {
		return err
	}

	ev, err := sg.Changesets.Update(cliCtx, &sourcegraph.ChangesetUpdateOp{
		Repo: repo.RepoSpec(),
		ID:   c.Args.ID,
		Open: true,
	})
	if err != nil {
		return err
	}

	log.Printf("# opened changeset %s #%d", repo.URI, ev.After.ID)
	return nil
}
