package main

import (
	"context"
	"net/url"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

type GithubCodeHost struct {
	def *CodeHostDefinition
	c   *github.Client
}

func NewGithubCodeHost(ctx context.Context, def *CodeHostDefinition) (*GithubCodeHost, error) {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: def.Token},
	))

	gh, err := github.NewEnterpriseClient(def.URL, def.URL, tc)
	if err != nil {
		return nil, err
	}
	return &GithubCodeHost{
		def: def,
		c:   gh,
	}, nil
}

func (g *GithubCodeHost) ListRepos(ctx context.Context) ([]*Repo, error) {
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

	res := make([]*Repo, 0, len(repos))
	for _, repo := range repos {
		u, err := url.Parse(repo.GetGitURL())
		if err != nil {
			return nil, err
		}
		u.User = url.UserPassword(g.def.Username, g.def.Password)
		u.Scheme = "https"
		res = append(res, &Repo{
			name: repo.GetName(),
			url:  u.String(),
		})
	}

	return res, nil
}
