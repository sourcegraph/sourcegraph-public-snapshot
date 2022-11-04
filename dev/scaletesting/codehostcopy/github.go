package main

import (
	"context"
	"net/url"
	"strings"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"github.com/sourcegraph/sourcegraph/dev/scaletesting/internal/store"
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
