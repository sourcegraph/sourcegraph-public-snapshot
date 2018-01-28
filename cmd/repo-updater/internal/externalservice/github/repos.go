package github

import (
	"encoding/json"
	"fmt"
	"strings"

	"context"

	"github.com/prometheus/client_golang/prometheus"
)

// SplitRepositoryNameWithOwner splits a GitHub repository's "owner/name" string into "owner" and "name", with
// validation.
func SplitRepositoryNameWithOwner(nameWithOwner string) (owner, repo string, err error) {
	parts := strings.SplitN(nameWithOwner, "/", 2)
	if len(parts) != 2 || strings.Contains(parts[1], "/") || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid GitHub repository \"owner/name\" string: %q", nameWithOwner)
	}
	return parts[0], parts[1], nil
}

// Repository is a GitHub repository.
type Repository struct {
	ID            string // ID of repository (GitHub GraphQL ID, not GitHub database ID)
	NameWithOwner string // full name of repository ("owner/name")
	Description   string // description of repository
	URL           string // the web URL of this repository ("https://github.com/foo/bar")
	IsFork        bool   // whether the repository is a fork of another repository
}

// RepositoryFieldsGraphQLFragment returns a GraphQL fragment that contains the fields needed to populate the
// Repository struct.
func (Repository) RepositoryFieldsGraphQLFragment() string {
	return `
fragment RepositoryFields on Repository {
	id
	nameWithOwner
	description
	url
	isFork
}
	`
}

// GetRepositoryMock is set by tests to mock (*Client).GetRepository.
var GetRepositoryMock func(ctx context.Context, owner, name string) (*Repository, error)

// MockGetRepository_Return is called by tests to mock (*Client).GetRepository.
func MockGetRepository_Return(returns *Repository) {
	GetRepositoryMock = func(context.Context, string, string) (*Repository, error) {
		return returns, nil
	}
}

// GetRepository gets a repository from GitHub by owner and repository name.
func (c *Client) GetRepository(ctx context.Context, owner, name string) (*Repository, error) {
	if GetRepositoryMock != nil {
		return GetRepositoryMock(ctx, owner, name)
	}

	key := owner + "/" + name

	if cached := c.getRepositoryFromCache(ctx, key); cached != nil {
		reposGitHubCacheCounter.WithLabelValues("hit").Inc()
		if cached.NotFound {
			return nil, ErrNotFound
		}
		return &cached.Repository, nil
	}

	repo, err := c.getRepositoryFromAPI(ctx, owner, name)
	if IsNotFound(err) {
		// Before we do anything, ensure we cache NotFound responses.
		// Do this if client is unauthed or authed, it's okay since we're only caching not found responses here.
		c.addRepositoryToCache(key, &cachedRepo{NotFound: true})
		reposGitHubCacheCounter.WithLabelValues("notfound").Inc()
	}
	if err != nil {
		reposGitHubCacheCounter.WithLabelValues("error").Inc()
		return nil, err
	}

	c.addRepositoryToCache(key, &cachedRepo{Repository: *repo})
	reposGitHubCacheCounter.WithLabelValues("miss").Inc()

	return repo, nil
}

var (
	reposGitHubCacheCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "src",
		Subsystem: "repos",
		Name:      "github_cache_hit",
		Help:      "Counts cache hits and misses for GitHub repo metadata.",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(reposGitHubCacheCounter)
}

type cachedRepo struct {
	Repository

	// NotFound indicates that the GitHub API reported that the repository was not found.
	NotFound bool
}

// getRepositoryFromCache attempts to get a response from the redis cache.
// It returns nil error for cache-hit condition and non-nil error for cache-miss.
func (c *Client) getRepositoryFromCache(ctx context.Context, key string) *cachedRepo {
	b, ok := c.repoCache.Get(strings.ToLower(key))
	if !ok {
		return nil
	}

	var cached cachedRepo
	if err := json.Unmarshal(b, &cached); err != nil {
		return nil
	}

	return &cached
}

// addRepositoryToCache will cache the value for repo.
func (c *Client) addRepositoryToCache(key string, repo *cachedRepo) {
	b, err := json.Marshal(repo)
	if err != nil {
		return
	}
	c.repoCache.Set(strings.ToLower(key), b)
}

// getRepositoryFromAPI attempts to fetch a repository from the GitHub API without use of the redis cache.
func (c *Client) getRepositoryFromAPI(ctx context.Context, owner, name string) (*Repository, error) {
	// If no token, we must use the older REST API, not the GraphQL API. See
	// https://platform.github.community/t/anonymous-access/2093/2. This situation occurs on (for
	// example) a server with autoAddRepos and no GitHub connection configured when someone visits
	// http://[sourcegraph-hostname]/github.com/foo/bar.
	//
	// To avoid having 2 code paths when getting a repo (REST API and GraphQL API), we just always
	// use the REST API.
	var result struct {
		ID          string `json:"node_id"`   // GraphQL ID
		FullName    string `json:"full_name"` // same as nameWithOwner
		Description string
		HTMLURL     string `json:"html_url"` // web URL
		Fork        bool
	}
	if err := c.requestGet(ctx, fmt.Sprintf("/repos/%s/%s", owner, name), &result); err != nil {
		return nil, err
	}
	return &Repository{
		ID:            result.ID,
		NameWithOwner: result.FullName,
		Description:   result.Description,
		URL:           result.HTMLURL,
		IsFork:        result.Fork,
	}, nil
}

// ListViewerRepositories lists GitHub repositories affiliated with the viewer (the currently authenticated user).
// The nextPageCursor is the ID value to pass back to this method (in the "after" parameter) to retrieve the next
// page of repositories.
func (c *Client) ListViewerRepositories(ctx context.Context, first int, after *string) (repos []*Repository, nextPageCursor *string, rateLimitCost int, err error) {
	var result struct {
		Viewer struct {
			Repositories struct {
				Nodes    []*Repository
				PageInfo struct {
					HasNextPage bool
					EndCursor   *string
				}
			}
		}
		RateLimit struct {
			Cost int
		}
	}
	if err := c.requestGraphQL(ctx, `
	query AffiliatedRepositories($first: Int!, $after: String) {
		viewer {
			repositories(
				first: $first
				after: $after
				affiliations: [OWNER, ORGANIZATION_MEMBER, COLLABORATOR]
				orderBy:{ field: PUSHED_AT, direction: DESC }
			) {
				nodes {
					...RepositoryFields
				}
				pageInfo {
					hasNextPage
					endCursor
				}
			}
		}
		rateLimit {
			cost
		}
	}`+(Repository{}).RepositoryFieldsGraphQLFragment(),
		map[string]interface{}{"first": first, "after": after},
		&result,
	); err != nil {
		return nil, nil, 0, err
	}

	// Add to cache.
	for _, repo := range result.Viewer.Repositories.Nodes {
		c.addRepositoryToCache(repo.NameWithOwner, &cachedRepo{Repository: *repo})
	}

	nextPageCursor = result.Viewer.Repositories.PageInfo.EndCursor
	if !result.Viewer.Repositories.PageInfo.HasNextPage {
		nextPageCursor = nil
	}
	return result.Viewer.Repositories.Nodes, nextPageCursor, result.RateLimit.Cost, nil
}
