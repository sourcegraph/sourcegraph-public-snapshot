package sgx

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os/exec"

	"strings"

	"sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sqs/pbtypes"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/sgx/cli"
	"src.sourcegraph.com/sourcegraph/util/timeutil"
)

func init() {
	changesetsGroup, err := cli.CLI.AddCommand("changeset",
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
}

type changesetsCmd struct{}

func (c *changesetsCmd) Execute(args []string) error { return nil }

type changesetListCmd struct {
	Repo   string `short:"r" long:"repo" description:"repository URI" required:"yes"`
	Status string `long:"status" description:"filter to only 'open' or 'closed' changesets (default: all)"`
}

func (c *changesetListCmd) Execute(args []string) error {
	cl := Client()

	if _, err := cl.Repos.Get(cliCtx, &sourcegraph.RepoSpec{URI: c.Repo}); err != nil {
		return err
	}

	for page := 1; ; page++ {
		changesets, err := cl.Changesets.List(cliCtx, &sourcegraph.ChangesetListOp{
			Repo:        c.Repo,
			Open:        c.Status == "open",
			Closed:      c.Status == "closed",
			ListOptions: sourcegraph.ListOptions{Page: int32(page)},
		})

		if err != nil {
			return err
		}
		if len(changesets.Changesets) == 0 {
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
	Repo  string `short:"r" long:"repo" description:"repository URI" required:"yes"`
	Base  string `long:"base" description:"base branch"`
	Head  string `long:"head" description:"head branch"`
	Title string `short:"t" long:"title" description:"title"`
}

func (c *changesetCreateCmd) Execute(args []string) error {
	cl := Client()

	conf, err := cl.Meta.Config(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}

	repo, err := cl.Repos.Get(cliCtx, &sourcegraph.RepoSpec{URI: c.Repo})
	if err != nil {
		return err
	}

	if c.Base == "" {
		c.Base = repo.DefaultBranch
	}
	if c.Base == "" {
		return errors.New("must specify --base (could not determine default branch for repo)")
	}
	if _, err := cl.Repos.GetCommit(cliCtx, &sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Base}); err != nil {
		return fmt.Errorf("checking that base branch %q exists on remote server: %s", c.Base, err)
	}

	if c.Head == "" {
		currentBranch, _ := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
		c.Head = strings.TrimSpace(string(currentBranch))
	}
	if c.Head == "" {
		return errors.New("must specify --head (could not determine current branch)")
	}
	remoteHeadCommit, err := cl.Repos.GetCommit(cliCtx, &sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Head})
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

	authInfo, err := cl.Auth.Identify(cliCtx, &pbtypes.Void{})
	if err != nil {
		return err
	}
	user, err := cl.Users.Get(cliCtx, &sourcegraph.UserSpec{UID: authInfo.UID})
	if err != nil {
		return err
	}

	title, description, err := newChangesetInEditor(c.Title)
	if err != nil {
		return err
	}

	changeset, err := cl.Changesets.Create(cliCtx, &sourcegraph.ChangesetCreateOp{
		Repo: sourcegraph.RepoSpec{URI: c.Repo},
		Changeset: &sourcegraph.Changeset{
			Title:       title,
			Description: description,
			Author:      user.Spec(),
			DeltaSpec: &sourcegraph.DeltaSpec{
				Base: sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Base},
				Head: sourcegraph.RepoRevSpec{RepoSpec: repo.RepoSpec(), Rev: c.Head},
			},
		},
	})
	if err != nil {
		return err
	}

	url, err := url.Parse(conf.AppURL)
	if err != nil {
		return err
	}
	log.Println(url.ResolveReference(router.Rel.URLToRepoChangeset(repo.URI, changeset.ID)))
	return nil
}

func newChangesetInEditor(origTitle string) (title, description string, err error) {
	contents := origTitle + `
# Please enter the changeset title (in the first line) and description
# (in the subsequent lines). Lines starting with '#' will be ignored,
# and an empty message aborts the changeset.
`

	txt, err := openTempFileInEditor([]byte(contents))
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
