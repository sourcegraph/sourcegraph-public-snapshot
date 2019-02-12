package repos

import (
	"context"
	"net/url"
	"sync/atomic"

	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

// A GithubSource yields repositories from multiple Github connections configured
// in Sourcegraph via the external services configuration and retrieved
// from the Frontend API.
type GithubSource struct {
	configs GithubConnConfigs
}

// A GithubConnConfigs returns configured Github connections to be used in
// GithubSource to access upstream repos.
type GithubConnConfigs func(context.Context) ([]*schema.GitHubConnection, error)

// NewGithubSource returns a new GithubSource with the given configs.
func NewGithubSource(configs GithubConnConfigs) *GithubSource {
	return &GithubSource{configs: configs}
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

func githubRepoToRepo(ghrepo *github.Repository, conn *githubConnection) *Repo {
	return &Repo{
		Name:         string(githubRepositoryToRepoPath(conn, ghrepo)),
		ExternalRepo: *github.ExternalRepoSpec(ghrepo, *conn.baseURL),
		Description:  ghrepo.Description,
		Fork:         ghrepo.IsFork,
		Archived:     ghrepo.IsArchived,
	}
}
