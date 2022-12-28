package gqltestutil

import "github.com/sourcegraph/sourcegraph/lib/errors"

// GitBlob returns blob content of the file in given repository at given revision.
func (c *Client) GitBlob(repoName, revision, filePath string) (string, error) {
	const gqlQuery = `
query Blob($repoName: String!, $revision: String!, $filePath: String!) {
	repository(name: $repoName) {
		commit(rev: $revision) {
			file(path: $filePath) {
				content
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
						Content string `json:"content"`
					} `json:"file"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"data"`
	}
	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository.Commit.File.Content, nil
}

// GitListFilenames lists all files for the repo
func (c *Client) GitListFilenames(repoName, revision string) ([]string, error) {
	const gqlQuery = `
query Files($repoName: String!, $revision: String!) {
	repository(name: $repoName) {
		commit(rev: $revision) {
            fileNames
		}
	}
}
`
	variables := map[string]any{
		"repoName": repoName,
		"revision": revision,
	}
	var resp struct {
		Data struct {
			Repository struct {
				Commit struct {
					FileNames []string `json:"fileNames"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"data"`
	}
	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository.Commit.FileNames, nil
}

// GitGetCommitMessage returns commit message for given repo and revision.
// This spins up sub-repo permissions for the commit and error is returned when
// trying to access restricted commit
func (c *Client) GitGetCommitMessage(repoName, revision string) (string, error) {
	const gqlQuery = `
query Files($repoName: String!, $revision: String!) {
	repository(name: $repoName) {
		commit(rev: $revision) {
            message
		}
	}
}
`
	variables := map[string]any{
		"repoName": repoName,
		"revision": revision,
	}
	var resp struct {
		Data struct {
			Repository struct {
				Commit struct {
					Message string `json:"message"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"data"`
	}
	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository.Commit.Message, nil
}

// GitGetCommitSymbols returns symbols of all files in a given commit.
func (c *Client) GitGetCommitSymbols(repoName, revision string) ([]SimplifiedSymbolNode, error) {
	const gqlQuery = `
query CommitSymbols($repoName: String!, $revision: String!) {
	repository(name: $repoName) {
		commit(rev: $revision) {
            symbols(query: "") {
				nodes {
					name
					kind
					location {
						resource {
							path
						}
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
	}
	var resp struct {
		Data struct {
			Repository struct {
				Commit struct {
					Symbols struct {
						Nodes []SimplifiedSymbolNode `json:"nodes"`
					} `json:"symbols"`
				} `json:"commit"`
			} `json:"repository"`
		} `json:"data"`
	}
	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Repository.Commit.Symbols.Nodes, nil
}

type SimplifiedSymbolNode struct {
	Name     string `json:"name"`
	Kind     string `json:"kind"`
	Location struct {
		Resource struct {
			Path string `json:"path"`
		} `json:"resource"`
	} `json:"location"`
}
