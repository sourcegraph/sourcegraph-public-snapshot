package repos

import (
	"context"
	"fmt"
	"net/url"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/jsonc"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A Source yields repositories to be stored and analysed by Sourcegraph.
// Successive calls to its ListRepos method may yield different results.
type Source interface {
	ListRepos(context.Context) ([]*Repo, error)
}

// A GithubSource yields repositories from multiple Github connections configured
// in Sourcegraph via the external services configuration and retrieved
// from the Frontend API.
type GithubSource struct {
	api InternalAPI
}

// NewGithubSource returns a new GithubSource with the given configs.
func NewGithubSource(api InternalAPI) *GithubSource {
	return &GithubSource{api: api}
}

// ListRepos returns all Github repositories accessible to all connections configured
// in Sourcegraph via the external services configuration.
func (s GithubSource) ListRepos(ctx context.Context) ([]*Repo, error) {
	conns, err := s.connections(ctx)
	if err != nil {
		log15.Error("unable to fetch GitHub connections", "err", err)
		return nil, err
	}

	type repository struct {
		conn *githubConnection
		repo *github.Repository
	}

	ch := make(chan repository)
	done := uint32(0)
	for _, c := range conns {
		go func(c *githubConnection) {
			for r := range c.listAllRepositories(ctx) {
				ch <- repository{conn: c, repo: r}
			}
			if atomic.AddUint32(&done, 1) == uint32(len(conns)) { // All done, close channel
				close(ch)
			}
		}(c)
	}

	var repos []*Repo
	for r := range ch {
		repos = append(repos, githubRepoToRepo(r.repo, r.conn))
	}

	return repos, nil
}

func (s GithubSource) connections(ctx context.Context) ([]*githubConnection, error) {
	configs, err := s.configs(ctx)
	if err != nil {
		return nil, err
	}

	var hasGitHubDotComConnection bool
	for _, c := range configs {
		u, _ := url.Parse(c.Url)
		if u != nil && (u.Hostname() == "github.com" || u.Hostname() == "www.github.com" || u.Hostname() == "api.github.com") {
			hasGitHubDotComConnection = true
			break
		}
	}

	if !hasGitHubDotComConnection {
		// Add a GitHub.com entry by default, to support navigating to URL paths like
		// /github.com/foo/bar to auto-add that repository.
		configs = append(configs, &schema.GitHubConnection{
			RepositoryQuery:             []string{"none"}, // don't try to list all repositories during syncs
			Url:                         "https://github.com",
			InitialRepositoryEnablement: true,
		})
	}

	conns := make([]*githubConnection, 0, len(configs))
	for _, c := range configs {
		conn, err := newGitHubConnection(c)
		if err != nil {
			log15.Error("Error processing configured GitHub connection. Skipping it.", "url", c.Url, "error", err)
			continue
		}
		conns = append(conns, conn)
	}

	return conns, nil
}

func (s GithubSource) configs(ctx context.Context) ([]*schema.GitHubConnection, error) {
	svcs, err := s.api.ExternalServicesList(ctx, api.ExternalServicesListRequest{Kind: "GITHUB"})
	if err != nil {
		return nil, err
	}

	configs := make([]*schema.GitHubConnection, len(svcs))
	for i, s := range svcs {
		if err := jsonc.Unmarshal(s.Config, &configs[i]); err != nil {
			return nil, fmt.Errorf("github source: config error: %s", err)
		}
	}

	return configs, nil
}

func githubRepoToRepo(ghrepo *github.Repository, conn *githubConnection) *Repo {
	return &Repo{
		Name:         string(githubRepositoryToRepoPath(conn, ghrepo)),
		CloneURL:     conn.authenticatedRemoteURL(ghrepo),
		ExternalRepo: *github.ExternalRepoSpec(ghrepo, *conn.baseURL),
		Description:  ghrepo.Description,
		Fork:         ghrepo.IsFork,
		Archived:     ghrepo.IsArchived,
	}
}
