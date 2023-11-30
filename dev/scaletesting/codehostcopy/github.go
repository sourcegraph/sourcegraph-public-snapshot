package main

import (
	"context"
	"fmt"
	"math"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/v55/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitHubCodeHost struct {
	def     *CodeHostDefinition
	c       *github.Client
	page    int
	perPage int
	done    bool
	err     error
}

var (
	_ CodeHostSource      = (*GitHubCodeHost)(nil)
	_ CodeHostDestination = (*GitHubCodeHost)(nil)
)

func NewGitHubCodeHost(ctx context.Context, def *CodeHostDefinition) (*GitHubCodeHost, error) {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: def.Token},
	))

	baseURL, err := url.Parse(def.URL)
	if err != nil {
		return nil, err
	}
	baseURL.Path = "/api/v3"

	gh, err := github.NewClient(tc).WithEnterpriseURLs(baseURL.String(), baseURL.String())
	if err != nil {
		return nil, errors.Wrap(err, "failed to create GitHub client")
	}
	return &GitHubCodeHost{
		def: def,
		c:   gh,
	}, nil
}

// GitOpts returns the options that should be used when a git command is invoked for Github
func (g *GitHubCodeHost) GitOpts() []GitOpt {
	if len(g.def.SSHKey) == 0 {
		return []GitOpt{}
	}

	GitEnv := func(cmd *run.Command) *run.Command {
		return cmd.Environ([]string{fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no'", g.def.SSHKey)})
	}

	return []GitOpt{GitEnv}
}

// AddSSHKey adds the SSH key defined in the code host configuration to
// the current authenticated user.
//
// If there is no ssh key defined on the code host configuration this
// is is a noop and returns a 0 for the key ID
func (g *GitHubCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	if len(g.def.SSHKey) == 0 {
		return 0, nil
	}
	data, err := os.ReadFile(g.def.SSHKey)
	if err != nil {
		return 0, err
	}

	keyData := string(data)
	keyTitle := "codehost-copy key"
	githubKey := github.Key{
		Key:   &keyData,
		Title: &keyTitle,
	}

	result, res, err := g.c.Users.CreateKey(ctx, &githubKey)
	if err != nil {
		return 0, err
	}
	if res.StatusCode >= 300 {
		return 0, errors.Newf("failed to add key. Got status %d code", res.StatusCode)
	}

	return *result.ID, nil
}

// DropSSHKey removes the ssh key by by ID for the current authenticated user. If there is no
// ssh key set on the codehost configuration this method is a noop
func (g *GitHubCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
	// if there is no ssh key in the code host definition
	// then we have nothing to drop
	if len(g.def.SSHKey) == 0 {
		return nil
	}
	res, err := g.c.Users.DeleteKey(ctx, keyID)
	if err != nil {
		return err
	}

	if res.StatusCode != 200 {
		return errors.Newf("failed to delete key %v. Got status %d code", keyID, res.StatusCode)
	}
	return nil
}

func (g *GitHubCodeHost) listRepos(ctx context.Context, start int, size int) ([]*store.Repo, int, error) {
	var repos []*github.Repository
	var resp *github.Response
	var err error
	var next int

	if strings.HasPrefix(g.def.Path, "@") {
		// If we're given a user and not an organization, query the user repos.
		opts := github.RepositoryListOptions{
			ListOptions: github.ListOptions{Page: start, PerPage: size},
		}
		repos, resp, err = g.c.Repositories.List(ctx, strings.Replace(g.def.Path, "@", "", 1), &opts)
		if err != nil {
			return nil, 0, err
		}

		if resp.StatusCode >= 300 {
			return nil, 0, errors.Newf("failed to list repos for user %s. Got status %d code", strings.Replace(g.def.Path, "@", "", 1), resp.StatusCode)
		}

		next = resp.NextPage
		// If next page is 0 we're at the last page, so set the last page
		if next == 0 && g.page != resp.LastPage {
			next = resp.LastPage
		}
	} else {
		opts := github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{Page: start, PerPage: size},
		}

		repos, resp, err = g.c.Repositories.ListByOrg(ctx, g.def.Path, &opts)
		if err != nil {
			return nil, 0, err
		}

		if resp.StatusCode >= 300 {
			return nil, 0, errors.Newf("failed to list repos for org %s. Got status %d code", g.def.Path, resp.StatusCode)
		}

		next = resp.NextPage
		// If next page is 0 we're at the last page, so set the last page
		if next == 0 && g.page != resp.LastPage {
			next = resp.LastPage
		}
	}

	res := make([]*store.Repo, 0, len(repos))
	for _, repo := range repos {
		u, err := url.Parse(repo.GetGitURL())
		if err != nil {
			return nil, 0, err
		}
		u.User = url.UserPassword(g.def.Username, g.def.Password)
		u.Scheme = "https"
		res = append(res, &store.Repo{
			Name:   repo.GetName(),
			GitURL: u.String(),
		})
	}

	return res, next, nil
}

func (g *GitHubCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	return nil, errors.New("not implemented")
}

func (g *GitHubCodeHost) Iterator() Iterator[[]*store.Repo] {
	return g
}

func (g *GitHubCodeHost) Next(ctx context.Context) []*store.Repo {
	if g.done {
		return nil
	}

	results, next, err := g.listRepos(ctx, g.page, g.perPage)
	if err != nil {
		g.err = err
		return nil
	}

	// when next is 0, it means the Github api returned the nextPage as 0, which indicates that there are not more pages to fetch
	if next > 0 {
		// Ensure that the next request starts at the next page
		g.page = next
	} else {
		g.done = true
	}

	return results
}

func (g *GitHubCodeHost) Done() bool {
	return g.done
}

func (g *GitHubCodeHost) Err() error {
	return g.err
}

func (g *GitHubCodeHost) getTotalPrivateRepos(ctx context.Context) (int, error) {
	// not supplied in the config, so get whatever GitHub tells us is present (but might be incorrect)
	if g.def.RepositoryLimit == 0 {
		if strings.HasPrefix(g.def.Path, "@") {
			u, resp, err := g.c.Users.Get(ctx, strings.Replace(g.def.Path, "@", "", 1))
			if err != nil {
				return 0, err
			}
			if resp.StatusCode >= 300 {
				return 0, errors.Newf("failed to get user %s. Got status %d code", strings.Replace(g.def.Path, "@", "", 1), resp.StatusCode)
			}

			return int(u.GetOwnedPrivateRepos()), nil
		} else {
			o, resp, err := g.c.Organizations.Get(ctx, g.def.Path)
			if err != nil {
				return 0, err
			}
			if resp.StatusCode >= 300 {
				return 0, errors.Newf("failed to get org %s. Got status %d code", g.def.Path, resp.StatusCode)
			}

			return int(o.GetOwnedPrivateRepos()), nil
		}
	} else {
		return g.def.RepositoryLimit, nil
	}
}

func (g *GitHubCodeHost) setPage(total int, remainder int) {
	// setting per page is not implemented yet so use GH default
	perPage := 10
	if g.perPage != 0 {
		perPage = g.perPage
	}
	g.page = int(math.Ceil(float64(total-remainder) / float64(perPage)))
}

func (g *GitHubCodeHost) InitializeFromState(ctx context.Context, stateRepos []*store.Repo) (int, int, error) {
	t, err := g.getTotalPrivateRepos(ctx)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "failed to get total private repos size for source %s", g.def.Path)
	}
	remainder := t - len(stateRepos)

	// Process started but not finished, set page to continue
	if len(stateRepos) != 0 && remainder != 0 {
		g.setPage(t, remainder)
	}

	return t, remainder, nil
}
