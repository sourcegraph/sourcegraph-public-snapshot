package repos

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	multierror "github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/pkg/api"
	"github.com/sourcegraph/sourcegraph/pkg/conf/reposource"
	"github.com/sourcegraph/sourcegraph/pkg/extsvc/github"
	"github.com/sourcegraph/sourcegraph/pkg/httpcli"
	"github.com/sourcegraph/sourcegraph/schema"
	log15 "gopkg.in/inconshreveable/log15.v2"
)

func githubRepositoryToRepoPath(conn *githubConnection, repo *github.Repository) api.RepoName {
	return reposource.GitHubRepoName(conn.config.RepositoryPathPattern, conn.originalHostname, repo.NameWithOwner)
}

func newGitHubConnection(config *schema.GitHubConnection, cf *httpcli.Factory) (*githubConnection, error) {
	baseURL, err := url.Parse(config.Url)
	if err != nil {
		return nil, err
	}
	baseURL = NormalizeBaseURL(baseURL)
	originalHostname := baseURL.Hostname()

	apiURL, githubDotCom := github.APIRoot(baseURL)

	if cf == nil {
		cf = NewHTTPClientFactory()
	}

	var opts []httpcli.Opt
	if config.Certificate != "" {
		pool, err := newCertPool(config.Certificate)
		if err != nil {
			return nil, err
		}
		opts = append(opts, httpcli.NewCertPoolOpt(pool))
	}

	cli, err := cf.Doer(opts...)
	if err != nil {
		return nil, err
	}

	exclude := make(map[string]bool, len(config.Exclude))
	for _, r := range config.Exclude {
		if r.Name != "" {
			exclude[strings.ToLower(r.Name)] = true
		}

		if r.Id != "" {
			exclude[r.Id] = true
		}
	}

	return &githubConnection{
		config:           config,
		exclude:          exclude,
		baseURL:          baseURL,
		githubDotCom:     githubDotCom,
		client:           github.NewClient(apiURL, config.Token, cli),
		searchClient:     github.NewClient(apiURL, config.Token, cli),
		originalHostname: originalHostname,
	}, nil
}

type githubConnection struct {
}

// authenticatedRemoteURL returns the repository's Git remote URL with the configured
// GitHub personal access token inserted in the URL userinfo.
func (c *githubConnection) authenticatedRemoteURL(repo *github.Repository) string {
	if c.config.GitURLType == "ssh" {
		url := fmt.Sprintf("git@%s:%s.git", c.originalHostname, repo.NameWithOwner)
		return url
	}

	if c.config.Token == "" {
		return repo.URL
	}
	u, err := url.Parse(repo.URL)
	if err != nil {
		log15.Warn("Error adding authentication to GitHub repository Git remote URL.", "url", repo.URL, "error", err)
		return repo.URL
	}
	u.User = url.User(c.config.Token)
	return u.String()
}

func (c *githubConnection) excludes(r *github.Repository) bool {
	return c.exclude[strings.ToLower(r.NameWithOwner)] || c.exclude[r.ID]
}

func (c *githubConnection) listAllRepositories(ctx context.Context) ([]*github.Repository, error) {
	set := make(map[int64]*github.Repository)
	errs := new(multierror.Error)

	for _, repositoryQuery := range c.config.RepositoryQuery {
		switch repositoryQuery {
		case "public":
			if c.githubDotCom {
				errs = multierror.Append(errs, errors.New(`unsupported configuration "public" for "repositoryQuery" for github.com`))
				continue
			}
			var sinceRepoID int64
			for {
				if err := ctx.Err(); err != nil {
					errs = multierror.Append(errs, err)
					break
				}

				repos, err := c.client.ListPublicRepositories(ctx, sinceRepoID)
				if err != nil {
					errs = multierror.Append(errs, errors.Wrapf(err, "failed to list public repositories: sinceRepoID=%d", sinceRepoID))
					break
				}
				if len(repos) == 0 {
					break
				}
				log15.Debug("github sync public", "repos", len(repos), "error", err)
				for _, r := range repos {
					set[r.DatabaseID] = r
					if sinceRepoID < r.DatabaseID {
						sinceRepoID = r.DatabaseID
					}
				}
			}
		case "affiliated":
			hasNextPage := true
			for page := 1; hasNextPage; page++ {
				if err := ctx.Err(); err != nil {
					errs = multierror.Append(errs, err)
					break
				}

				var repos []*github.Repository
				var rateLimitCost int
				var err error
				repos, hasNextPage, rateLimitCost, err = c.client.ListUserRepositories(ctx, page)
				if err != nil {
					errs = multierror.Append(errs, errors.Wrapf(err, "failed to list affiliated GitHub repositories page %d", page))
					break
				}
				rateLimitRemaining, rateLimitReset, rateLimitRetry, _ := c.client.RateLimit.Get()
				log15.Debug(
					"github sync: ListUserRepositories",
					"repos", len(repos),
					"rateLimitCost", rateLimitCost,
					"rateLimitRemaining", rateLimitRemaining,
					"rateLimitReset", rateLimitReset,
					"retryAfter", rateLimitRetry,
				)

				for _, r := range repos {
					if c.githubDotCom && r.IsFork && r.ViewerPermission == "READ" {
						log15.Debug("not syncing readonly fork", "repo", r.NameWithOwner)
						continue
					}
					set[r.DatabaseID] = r
				}

				if hasNextPage {
					time.Sleep(c.client.RateLimit.RecommendedWaitForBackgroundOp(rateLimitCost))
				}
			}

		case "none":
			// nothing to do

		default:
			// Run the query as a GitHub advanced repository search
			// (https://github.com/search/advanced).
			hasNextPage := true
			for page := 1; hasNextPage; page++ {
				if err := ctx.Err(); err != nil {
					errs = multierror.Append(errs, err)
					break
				}

				reposPage, err := c.searchClient.ListRepositoriesForSearch(ctx, repositoryQuery, page)
				if err != nil {
					errs = multierror.Append(errs, errors.Wrapf(err, "failed to list GitHub repositories for search: page=%q, searchString=%q,", page, repositoryQuery))
					break
				}

				if reposPage.TotalCount > 1000 {
					// GitHub's advanced repository search will only
					// return 1000 results. We specially handle this case
					// to ensure the admin gets a detailed error
					// message. https://github.com/sourcegraph/sourcegraph/issues/2562
					errs = multierror.Append(errs, errors.Errorf(`repositoryQuery %q would return %d results. GitHub's Search API only returns up to 1000 results. Please adjust your repository query into multiple queries such that each returns less than 1000 results. For example: {"repositoryQuery": %s}`, repositoryQuery, reposPage.TotalCount, exampleRepositoryQuerySplit(repositoryQuery)))
					break
				}

				hasNextPage = reposPage.HasNextPage
				repos := reposPage.Repos

				rateLimitRemaining, rateLimitReset, rateLimitRetry, _ := c.searchClient.RateLimit.Get()
				log15.Debug(
					"github sync: ListRepositoriesForSearch",
					"searchString", repositoryQuery,
					"repos", len(repos),
					"rateLimitRemaining", rateLimitRemaining,
					"rateLimitReset", rateLimitReset,
					"retryAfter", rateLimitRetry,
				)

				for _, r := range repos {
					set[r.DatabaseID] = r
				}

				if hasNextPage {
					// GitHub search has vastly different rate limits to
					// the normal GitHub API (30req/m vs
					// 5000req/h). RecommendedWaitForBackgroundOp has
					// heuristics tuned for the normal API, part of which
					// is to not sleep if we have ample rate limit left.
					//
					// So we only let the heuristic kick in if we have
					// less than 5 requests left.
					remaining, _, retryAfter, ok := c.searchClient.RateLimit.Get()
					if retryAfter > 0 || (ok && remaining < 5) {
						time.Sleep(c.searchClient.RateLimit.RecommendedWaitForBackgroundOp(1))
					}
				}
			}
		}
	}

	for _, nameWithOwner := range c.config.Repos {
		if err := ctx.Err(); err != nil {
			errs = multierror.Append(errs, err)
			break
		}

		owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
		if err != nil {
			errs = multierror.Append(errs, errors.New("Invalid GitHub repository: nameWithOwner="+nameWithOwner))
			break
		}
		repo, err := c.client.GetRepository(ctx, owner, name)
		if err != nil {
			// TODO(tsenart): When implementing dry-run, reconsider alternatives to return
			// 404 errors on external service config validation.
			if github.IsNotFound(err) {
				log15.Warn("skipping missing github.repos entry:", "name", nameWithOwner, "err", err)
				continue
			}
			errs = multierror.Append(errs, errors.Wrapf(err, "Error getting GitHub repository: nameWithOwner=%s", nameWithOwner))
			break
		}
		log15.Debug("github sync: GetRepository", "repo", repo.NameWithOwner)
		set[repo.DatabaseID] = repo
		time.Sleep(c.client.RateLimit.RecommendedWaitForBackgroundOp(1)) // 0-duration sleep unless nearing rate limit exhaustion
	}

	repos := make([]*github.Repository, 0, len(set))
	for _, repo := range set {
		if !c.excludes(repo) {
			repos = append(repos, repo)
		}
	}

	return repos, errs.ErrorOrNil()
}

func (c *githubConnection) getRepository(ctx context.Context, nameWithOwner string) (*github.Repository, error) {
	owner, name, err := github.SplitRepositoryNameWithOwner(nameWithOwner)
	if err != nil {
		return nil, errors.Wrapf(err, "Invalid GitHub repository: nameWithOwner="+nameWithOwner)
	}

	repo, err := c.client.GetRepository(ctx, owner, name)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

func exampleRepositoryQuerySplit(q string) string {
	var qs []string
	for _, suffix := range []string{"created:>=2019", "created:2018", "created:2016..2017", "created:<2016"} {
		qs = append(qs, fmt.Sprintf("%s %s", q, suffix))
	}
	// Avoid escaping < and >
	var b bytes.Buffer
	enc := json.NewEncoder(&b)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(qs)
	return strings.TrimSpace(b.String())
}
