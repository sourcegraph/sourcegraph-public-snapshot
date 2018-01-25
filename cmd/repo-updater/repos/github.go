package repos

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	log15 "gopkg.in/inconshreveable/log15.v2"
	"sourcegraph.com/sourcegraph/sourcegraph/cmd/repo-updater/internal/externalservice/github"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/api"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/schema"
)

// RunGitHubRepositorySyncWorker runs the worker that syncs repositories from the configured GitHub and GitHub
// Enterprise instances to Sourcegraph.
func RunGitHubRepositorySyncWorker(ctx context.Context) error {
	var clients []*githubClient
	for _, c := range conf.Get().Github {
		client, err := newGitHubClient(c)
		if err != nil {
			return fmt.Errorf("error processing GitHub config %s: %s", c.Url, err)
		}
		clients = append(clients, client)
	}

	if len(clients) == 0 {
		return nil
	}
	for _, c := range clients {
		go func(c *githubClient) {
			for {
				updateGitHubRepositories(ctx, c)
				time.Sleep(updateInterval)
			}
		}(c)
	}
	select {}
}

// updateGitHubRepositories ensures that all provided repositories have been added and updated on Sourcegraph.
func updateGitHubRepositories(ctx context.Context, client *githubClient) {
	repos := client.listAllRepositories(ctx)

	githubRepositoryToRepoPath := func(repositoryPathPattern string, repo *github.Repository) api.RepoURI {
		if repositoryPathPattern == "" {
			repositoryPathPattern = "{host}/{nameWithOwner}"
		}
		return api.RepoURI(strings.NewReplacer(
			"{host}", client.originalHostname,
			"{nameWithOwner}", repo.NameWithOwner,
		).Replace(repositoryPathPattern))
	}

	repoChan := make(chan api.RepoCreateOrUpdateRequest)
	go createEnableUpdateRepos(ctx, nil, repoChan)
	for repo := range repos {
		log15.Debug("github sync: create/enable/update repo", "repo", repo.NameWithOwner)
		repoChan <- api.RepoCreateOrUpdateRequest{
			RepoURI:     githubRepositoryToRepoPath(client.config.RepositoryPathPattern, repo),
			Description: repo.Description,
			Fork:        repo.IsFork,
			Enabled:     client.config.InitialRepositoryEnablement,
		}
	}
	close(repoChan)
}

func newGitHubClient(config schema.GitHubConnection) (*githubClient, error) {
	baseURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	originalHostname := baseURL.Hostname()

	// GitHub.com's API is hosted on api.github.com.
	if hostname := strings.ToLower(baseURL.Hostname()); hostname == "github.com" || hostname == "www.github.com" {
		baseURL.Scheme = "https"
		baseURL.Host = "api.github.com" // might even be changed to http://github-proxy, but the github pkg handles that
	}

	var transport http.RoundTripper
	if config.Certificate != "" {
		var err error
		transport, err = transportWithCertTrusted(config.Certificate)
		if err != nil {
			return nil, err
		}
	}

	return &githubClient{
		config:           config,
		originalHostname: originalHostname,
		client:           github.NewClient(baseURL, config.Token, transport),
	}, nil
}

type githubClient struct {
	config schema.GitHubConnection
	client *github.Client

	// originalHostname is the hostname of config.Url (differs from client baseURL, whose host is api.github.com
	// for an originalHostname of github.com).
	originalHostname string
}

func (c *githubClient) listAllRepositories(ctx context.Context) <-chan *github.Repository {
	const first = 100 // max GitHub API "first" parameter
	ch := make(chan *github.Repository, first)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		if len(c.config.RepositoryQuery) == 0 {
			// Users need to specify ["none"] to disable affiliated default.
			c.config.RepositoryQuery = []string{"affiliated"}
		}
		for _, repositoryQuery := range c.config.RepositoryQuery {
			switch repositoryQuery {
			case "affiliated":
				// Avoid overloading the GitHub API. Even if the client has its own rate limiter, the sync worker
				// here should use an even more conservative rate limiter because it is a background job.
				rateLimiter := time.NewTicker(5 * time.Second)
				defer rateLimiter.Stop()

				var endCursor *string // GraphQL pagination cursor
				for {
					var repos []*github.Repository
					var err error
					repos, endCursor, err = c.client.ListViewerRepositories(ctx, first, endCursor)
					if err != nil {
						log15.Error("Error listing viewer's affiliated GitHub repositories", "endCursor", endCursor, "error", err)
						break
					}
					log15.Debug("github sync: ListViewerRepositories", "repos", len(repos))
					for _, r := range repos {
						log15.Debug("github sync: ListViewerRepositories: repo", "repo", r.NameWithOwner)
						ch <- r
					}
					if endCursor == nil {
						break
					}
					<-rateLimiter.C
				}

			case "none":
				// nothing to do

			default:
				log15.Error("Skipping unrecognized GitHub configuration repositoryQuery", "repositoryQuery", repositoryQuery)
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, nameWithOwner := range c.config.Repos {
			owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
			if err != nil {
				log15.Error("Invalid GitHub repository", "nameWithOwner", nameWithOwner)
				continue
			}
			repo, err := c.client.GetRepository(ctx, owner, name)
			if err != nil {
				log15.Error("Error getting GitHub repository", "nameWithOwner", nameWithOwner, "error", err)
				continue
			}
			log15.Debug("github sync: GetRepository", "repo", repo.NameWithOwner)
			ch <- repo
		}
	}()

	go func() {
		wg.Wait()
		close(ch)
	}()

	return ch
}
