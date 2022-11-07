package main

import (
	"context"
	"fmt"
	"net/url"
	"os"

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

func (g *GithubCodeHost) GitOpts() []GitOpt {
	if len(g.def.SSHKey) == 0 {
		return []GitOpt{}
	}

	GitEnv := func(cmd *run.Command) *run.Command {
		return cmd.Environ([]string{fmt.Sprintf("GIT_SSH_COMMAND=ssh -i %s -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no'", g.def.SSHKey)})
	}

	return []GitOpt{GitEnv}

}

func (g *GithubCodeHost) AddSSHKey(ctx context.Context) (int64, error) {
	if len(g.def.SSHKey) == 0 {
		return 0, nil
	}
	data, err := os.ReadFile(g.def.SSHKey)
	if err != nil {
		return 0, err
	}

	keyData := string(data)
	keyTitle := "CodeHostCopy key"
	githubKey := github.Key{
		Key:   &keyData,
		Title: &keyTitle,
	}

	result, res, err := g.c.Users.CreateKey(ctx, &githubKey)
	if err != nil {
		return 0, err
	}
	if res.StatusCode != 200 {
		return 0, errors.Newf("failed to add key. Got status %d code", res.StatusCode)
	}

	return *result.ID, nil
}

func (g *GithubCodeHost) DropSSHKey(ctx context.Context, keyID int64) error {
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
	opts := github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{},
	}
	var repos []*github.Repository
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
