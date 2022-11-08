package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/run"

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GithubCodeHost struct {
	def *CodeHostDefinition
	c   *github.Client
}

var _ CodeHostSource = (*GithubCodeHost)(nil)
var _ CodeHostDestination = (*GithubCodeHost)(nil)

func NewGithubCodeHost(ctx context.Context, def *CodeHostDefinition) (*GithubCodeHost, error) {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: def.Token},
	))

	baseURL, err := url.Parse(def.URL)
	if err != nil {
		return nil, err
	}
	baseURL.Path = "/api/v3"

	gh, err := github.NewEnterpriseClient(baseURL.String(), baseURL.String(), tc)
	if err != nil {
		return nil, err
	}
	return &GithubCodeHost{
		def: def,
		c:   gh,
	}, nil
}

// GitOpts returns the options that should be used when a git command is invoked for Github
func (g *GithubCodeHost) GitOpts() []GitOpt {
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
func (g *GithubCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
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
func (g *GithubCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
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

func (g *GithubCodeHost) ListRepos(ctx context.Context) ([]*store.Repo, error) {
	var repos []*github.Repository

	if strings.HasPrefix(g.def.Path, "@") {
		// If we're given a user and not an organization, query the user repos.
		opts := github.RepositoryListOptions{
			ListOptions: github.ListOptions{},
		}
		for {
			rs, resp, err := g.c.Repositories.List(ctx, strings.Replace(g.def.Path, "@", "", 1), &opts)
			if err != nil {
				return nil, err
			}
			repos = append(repos, rs...)

			if resp.NextPage == 0 {
				break
			}
			opts.ListOptions.Page = resp.NextPage
		}
	} else {
		opts := github.RepositoryListByOrgOptions{
			ListOptions: github.ListOptions{},
		}
		for {
			rs, resp, err := g.c.Repositories.ListByOrg(ctx, g.def.Path, &opts)
			if err != nil {
				return nil, err
			}
			repos = append(repos, rs...)

			if resp.NextPage == 0 {
				break
			}
			opts.ListOptions.Page = resp.NextPage
		}
	}

	res := make([]*store.Repo, 0, len(repos))
	for _, repo := range repos {
		u, err := url.Parse(repo.GetGitURL())
		if err != nil {
			return nil, err
		}
		u.User = url.UserPassword(g.def.Username, g.def.Password)
		u.Scheme = "https"
		res = append(res, &store.Repo{
			Name:   repo.GetName(),
			GitURL: u.String(),
		})
	}

	return res, nil
}

func (g *GithubCodeHost) CreateRepo(ctx context.Context, name string) (*url.URL, error) {
	return nil, errors.New("not implemented")
}
