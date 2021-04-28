package gqltestutil

import (
	"github.com/pkg/errors"
)

type CreateSearchContextInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
}

type SearchContextRepositoryRevisionsInput struct {
	RepositoryID string   `json:"repositoryID"`
	Revisions    []string `json:"revisions"`
}

// CreateSearchContext creates a new search context with the given input and repository revisions to be searched.
// It returns the GraphQL node ID of the newly created search context.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) CreateSearchContext(input CreateSearchContextInput, repositories []SearchContextRepositoryRevisionsInput) (string, error) {
	const query = `
mutation CreateSearchContext($input: SearchContextInput!, $repositories: [SearchContextRepositoryRevisionsInput!]!) {
	createSearchContext(searchContext: $input, repositories: $repositories) {
		id
	}
}
`
	variables := map[string]interface{}{
		"input":        input,
		"repositories": repositories,
	}
	var resp struct {
		Data struct {
			CreateSearchContext struct {
				ID string `json:"id"`
			} `json:"createSearchContext"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.CreateSearchContext.ID, nil
}

// DeleteSearchContext deletes a search context with the given id.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) DeleteSearchContext(id string) error {
	const query = `
mutation DeleteSearchContext($id: ID!) {
	 deleteSearchContext(id: $id) {
		alwaysNil
	}
}
`
	variables := map[string]interface{}{
		"id": id,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}
