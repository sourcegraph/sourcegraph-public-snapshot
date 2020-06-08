package e2eutil

import (
	"github.com/pkg/errors"
)

type SearchRepositoryResult struct {
	Name string `json:"name"`
}

// SearchRepositories search repositories with given query.
func (c *Client) SearchRepositories(query string) ([]*SearchRepositoryResult, error) {
	const gqlQuery = `query Search($query: String!) {
	search(query: $query) {
		results {
			results {
				... on Repository {
					name
				}
			}
		}
	}
}
`
	variables := map[string]interface{}{
		"query": query,
	}
	var resp struct {
		Data struct {
			Search struct {
				Results struct {
					Results []*SearchRepositoryResult `json:"results"`
				} `json:"results"`
			} `json:"search"`
		} `json:"data"`
	}
	err := c.GraphQL("", gqlQuery, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.Search.Results.Results, nil
}
