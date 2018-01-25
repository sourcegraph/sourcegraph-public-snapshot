package github

import (
	"encoding/json"
	"errors"
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
			return nil, errNotFound
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

var errNotFound = errors.New("GitHub repository not found")

// IsNotFound reports whether err is a GitHub API error of type NOT_FOUND or the equivalent cached response error.
func IsNotFound(err error) bool {
	if err == errNotFound {
		return true
	}
	errs, ok := err.(graphqlErrors)
	if !ok {
		return false
	}
	for _, err := range errs {
		if err.Type == "NOT_FOUND" {
			return true
		}
	}
	return false
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
	var result struct {
		Repository *Repository `json:"repository"`
	}
	if err := c.requestGraphQL(ctx, `
query Repository($owner: String!, $name: String!) {
	repository(owner: $owner, name: $name) {
		...RepositoryFields
	}
}`+(Repository{}).RepositoryFieldsGraphQLFragment(),
		map[string]interface{}{"owner": owner, "name": name},
		&result,
	); err != nil {
		return nil, err
	}
	if result.Repository == nil {
		return nil, errors.New("repository not found")
	}
	return result.Repository, nil
}

// ListViewerRepositories lists GitHub repositories affiliated with the viewer (the currently authenticated user).
// The nextPageCursor is the ID value to pass back to this method (in the "after" parameter) to retrieve the next
// page of repositories.
func (c *Client) ListViewerRepositories(ctx context.Context, first int, after *string) (repos []*Repository, nextPageCursor *string, err error) {
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
	}`+(Repository{}).RepositoryFieldsGraphQLFragment(),
		map[string]interface{}{"first": first, "after": after},
		&result,
	); err != nil {
		return nil, nil, err
	}

	// Add to cache.
	for _, repo := range result.Viewer.Repositories.Nodes {
		c.addRepositoryToCache(repo.NameWithOwner, &cachedRepo{Repository: *repo})
	}

	nextPageCursor = result.Viewer.Repositories.PageInfo.EndCursor
	if !result.Viewer.Repositories.PageInfo.HasNextPage {
		nextPageCursor = nil
	}
	return result.Viewer.Repositories.Nodes, nextPageCursor, nil
}
