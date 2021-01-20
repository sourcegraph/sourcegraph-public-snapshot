package gqltestutil

import (
	"context"
	"time"

	"github.com/pkg/errors"
)

// WaitForReposToBeCloned waits (up to two minutes) for all repositories
// in the list to be cloned.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) WaitForReposToBeCloned(repos ...string) error {
	timeout := 120 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var name string
	var missing []string
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("timed out in %s, still missing %v", timeout, missing)
		default:
		}

		const query = `
query Repositories {
	repositories(first: 1000, cloned: true, notCloned: false) {
		nodes {
			name
		}
	}
}
`
		var err error
		missing, err = c.waitForReposByQuery(name, query, repos...)
		if err != nil {
			return errors.Wrap(err, "wait for repos")
		}
		if len(missing) == 0 {
			break
		}

		// We want to log the very fist query of this kind, but don't want to create log spam
		// for subsequent queries.
		if name == "" {
			name = "WaitForReposToBeCloned"
		}

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

// WaitForReposToBeIndex waits (up to 30 seconds) for all repositories
// in the list to be indexed.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) WaitForReposToBeIndex(repos ...string) error {
	timeout := 180 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var name string
	var missing []string
	for {
		select {
		case <-ctx.Done():
			return errors.Errorf("timed out in %s, still missing %v", timeout, missing)
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
		missing, err = c.waitForReposByQuery(name, query, repos...)
		if err != nil {
			return errors.Wrap(err, "wait for repos")
		}
		if len(missing) == 0 {
			break
		}

		// We want to log the very fist query of this kind, but don't want to create log spam
		// for subsequent queries.
		if name == "" {
			name = "WaitForReposToBeIndex"
		}

		time.Sleep(100 * time.Millisecond)
	}
	return nil
}

func (c *Client) waitForReposByQuery(name, query string, repos ...string) ([]string, error) {
	var resp struct {
		Data struct {
			Repositories struct {
				Nodes []struct {
					Name string `json:"name"`
				} `json:"nodes"`
			} `json:"repositories"`
		} `json:"data"`
	}
	err := c.GraphQL(name, "", query, nil, &resp)
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
						serviceType
					}
				}
			}
		}
	}
}
`
	variables := map[string]interface{}{
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
	err := c.GraphQL("", "", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository.Commit.File.ExternalURLs, nil
}
