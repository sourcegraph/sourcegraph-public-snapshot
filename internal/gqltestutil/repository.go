package gqltestutil

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// WaitForReposToBeCloned waits up to two minutes for all repositories
// in the list to be cloned.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) WaitForReposToBeCloned(repos ...string) error {
	timeout := 120 * time.Second
	return c.WaitForReposToBeClonedWithin(timeout, repos...)
}

// WaitForReposToBeClonedWithin waits up to specified duration for all
// repositories in the list to be cloned.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) WaitForReposToBeClonedWithin(timeout time.Duration, repos ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var missing []string
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("wait for repos to be cloned timed out in %s, still missing %v", timeout, missing)
		default:
		}

		const query = `
query Repositories {
	repositories(first: 1000, cloneStatus: CLONED) {
		nodes {
			name
		}
	}
}
`
		var err error
		missing, err = c.waitForReposByQuery(query, repos...)
		if err != nil {
			return errors.Wrap(err, "wait for repos to be cloned")
		}
		if len(missing) == 0 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// DeleteRepoFromDiskByName will remove the repo form disk on GitServer.
func (c *Client) DeleteRepoFromDiskByName(name string) error {
	repo, err := c.Repository(name)
	if err != nil {
		return errors.Wrap(err, "getting repo")
	}
	if repo == nil {
		// Repo doesn't exist, no point trying to delete it
		return nil
	}

	q := fmt.Sprintf(`
mutation {
  deleteRepositoryFromDisk(repo:"%s") {
    alwaysNil
  }
}
`, repo.ID)

	err = c.GraphQL("", q, nil, nil)
	return errors.Wrap(err, "deleting repo from disk")
}

// WaitForReposToBeIndexed waits (up to 30 seconds) for all repositories
// in the list to be indexed.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) WaitForReposToBeIndexed(repos ...string) error {
	timeout := 180 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var missing []string
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("wait for repos to be indexed timed out in %s, still missing %v", timeout, missing)
		default:
		}

		const query = `
query Repositories {
	repositories(first: 1000, notIndexed: false, notCloned: false) {
		nodes {
			name
		}
	}
}
`
		var err error
		missing, err = c.waitForReposByQuery(query, repos...)
		if err != nil {
			return errors.Wrap(err, "wait for repos to be indexed")
		}
		if len(missing) == 0 {
			break
		}

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (c *Client) waitForReposByQuery(query string, repos ...string) ([]string, error) {
	var resp struct {
		Data struct {
			Repositories struct {
				Nodes []struct {
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	repoSet := make(map[string]struct{}, len(repos))
	for _, repo := range repos {
		repoSet[repo] = struct{}{}
	}
	for _, node := range resp.Data.Repositories.Nodes {
		delete(repoSet, node.Name)
	}
	if len(repoSet) > 0 {
		missing := make([]string, 0, len(repoSet))
		for name := range repoSet {
			missing = append(missing, name)
		}
		return missing, nil
	}

	return nil, nil
}

// ExternalLink is a link to an external service.
type ExternalLink struct {
	URL         string `json:"url"`         // The URL to the resource
	ServiceKind string `json:"serviceKind"` // The kind of service that the URL points to
	ServiceType string `json:"serviceType"` // The type of service that the URL points to
}

// FileExternalLinks external links for a file or directory in a repository.
func (c *Client) FileExternalLinks(repoName, revision, filePath string) ([]*ExternalLink, error) {
	const query = `
query FileExternalLinks($repoName: String!, $revision: String!, $filePath: String!) {
	repository(name: $repoName) {
		commit(rev: $revision) {
			file(path: $filePath) {
				externalURLs {
					... on ExternalLink {
						url
						serviceKind
						serviceType
					}
				}
			}
		}
	}
}
`
	variables := map[string]any{
		"repoName": repoName,
		"revision": revision,
		"filePath": filePath,
	}
	var resp struct {
		Data struct {
			Repository struct {
				Commit struct {
					File struct {
						ExternalURLs []*ExternalLink `json:"externalURLs"`
					} `json:"file"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository.Commit.File.ExternalURLs, nil
}

// Repository contains basic information of a repository from GraphQL.
type Repository struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// Repository returns basic information of the given repository.
func (c *Client) Repository(name string) (*Repository, error) {
	const query = `
query Repository($name: String!) {
	repository(name: $name) {
		id
		url
	}
}
`
	variables := map[string]any{
		"name": name,
	}
	var resp struct {
		Data struct {
			*Repository `json:"repository"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository, nil
}

// PermissionsInfo contains permissions information of a repository from
// GraphQL.
type PermissionsInfo struct {
	SyncedAt     time.Time
	UpdatedAt    time.Time
	Permissions  []string
	Unrestricted bool
}

// RepositoryPermissionsInfo returns permissions information of the given
// repository.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) RepositoryPermissionsInfo(name string) (*PermissionsInfo, error) {
	const query = `
query RepositoryPermissionsInfo($name: String!) {
	repository(name: $name) {
		permissionsInfo {
			syncedAt
			updatedAt
			permissions
			unrestricted
		}
	}
}
`
	variables := map[string]any{
		"name": name,
	}
	var resp struct {
		Data struct {
			Repository struct {
				*PermissionsInfo `json:"permissionsInfo"`
			} `json:"repository"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository.PermissionsInfo, nil
}

func (c *Client) AddRepoMetadata(repo string, key string, value *string) error {
	const query = `
mutation AddRepoMetadata($repo: ID!, $key: String!, $value: String) {
	addRepoMetadata(repo: $repo, key: $key, value: $value) {
		alwaysNil
	}
}
`
	variables := map[string]any{
		"repo":  repo,
		"key":   key,
		"value": value,
	}
	var resp map[string]interface{}
	return c.GraphQL("", query, variables, &resp)
}

func (c *Client) SetFeatureFlag(name string, value bool) error {
	const query = `
mutation SetFeatureFlag($name: String!, $value: Boolean!) {
	createFeatureFlag(name: $name, value: $value) {
		__typename
	}
}
`
	variables := map[string]any{
		"name":  name,
		"value": value,
	}
	var resp map[string]interface{}
	return c.GraphQL("", query, variables, &resp)
}
