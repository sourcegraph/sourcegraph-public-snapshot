package gqltestutil

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type CreateSearchContextInput struct {
	Name        string  `json:"name"`
	Namespace   *string `json:"namespace"`
	Description string  `json:"description"`
	Public      bool    `json:"public"`
	Query       string  `json:"query"`
}

type UpdateSearchContextInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Public      bool   `json:"public"`
	Query       string `json:"query"`
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
	variables := map[string]any{
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

type GetSearchContextResult struct {
	ID           string `json:"id"`
	Description  string `json:"description"`
	Spec         string `json:"spec"`
	AutoDefined  bool   `json:"autoDefined"`
	Repositories []struct {
		Repository struct {
			Name string `json:"name"`
		} `json:"repository"`
		Revisions []string `json:"revisions"`
	} `json:"repositories"`
	Query string `json:"query"`
}

func (c *Client) GetSearchContext(id string) (*GetSearchContextResult, error) {
	const query = `
query GetSearchContext($id: ID!) {
	node(id: $id) {
		... on SearchContext {
			id
			description
			spec
			autoDefined
			repositories {
				repository{
					name
				}
				revisions
			}
			query
		}
	}
}
`
	variables := map[string]any{
		"id": id,
	}
	var resp struct {
		Data struct {
			Node GetSearchContextResult `json:"node"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return &resp.Data.Node, nil
}

func (c *Client) UpdateSearchContext(id string, input UpdateSearchContextInput, repos []SearchContextRepositoryRevisionsInput) (string, error) {
	const query = `
mutation UpdateSearchContext($id: ID!, $input: SearchContextEditInput!, $repositories: [SearchContextRepositoryRevisionsInput!]!) {
	updateSearchContext(id: $id, searchContext: $input, repositories: $repositories) {
		id
		description
		spec
		autoDefined
		repositories {
			repository {
				name
			}
			revisions
		}
		query
	}
}
`
	variables := map[string]any{
		"id":           id,
		"input":        input,
		"repositories": repos,
	}
	var resp struct {
		Data struct {
			UpdateSearchContext GetSearchContextResult `json:"updateSearchContext"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	return resp.Data.UpdateSearchContext.ID, nil
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
	variables := map[string]any{
		"id": id,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

type SearchContextsOrderBy string

const (
	SearchContextsOrderByUpdatedAt SearchContextsOrderBy = "SEARCH_CONTEXT_UPDATED_AT"
	SearchContextsOrderBySpec      SearchContextsOrderBy = "SEARCH_CONTEXT_SPEC"
)

type ListSearchContextsOptions struct {
	First      int32                  `json:"first"`
	After      *string                `json:"after"`
	Query      *string                `json:"query"`
	Namespaces []*string              `json:"namespaces"`
	OrderBy    *SearchContextsOrderBy `json:"orderBy"`
	Descending bool                   `json:"descending"`
}

type ListSearchContextsResult struct {
	TotalCount int32 `json:"totalCount"`
	PageInfo   struct {
		HasNextPage bool    `json:"hasNextPage"`
		EndCursor   *string `json:"endCursor"`
	} `json:"pageInfo"`
	Nodes []GetSearchContextResult `json:"nodes"`
}

// ListSearchContexts list search contexts filtered by the given options.
func (c *Client) ListSearchContexts(options ListSearchContextsOptions) (*ListSearchContextsResult, error) {
	const query = `
query ListSearchContexts(
	$first: Int!
	$after: String
	$query: String
	$namespaces: [ID]
	$orderBy: SearchContextsOrderBy
	$descending: Boolean
) {
	searchContexts(
		first: $first
		after: $after
		query: $query
		namespaces: $namespaces
		orderBy: $orderBy
		descending: $descending
	) {
		nodes {
			id
			description
			spec
			autoDefined
			repositories {
				repository {
					name
				}
				revisions
			}
			query
		}
		pageInfo {
			hasNextPage
			endCursor
		}
		totalCount
	}
}`

	orderBy := SearchContextsOrderBySpec
	if options.OrderBy != nil {
		orderBy = *options.OrderBy
	}

	variables := map[string]any{
		"first":      options.First,
		"after":      options.After,
		"query":      options.Query,
		"namespaces": options.Namespaces,
		"orderBy":    orderBy,
		"descending": options.Descending,
	}

	var resp struct {
		Data struct {
			SearchContexts ListSearchContextsResult `json:"searchContexts"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	return &resp.Data.SearchContexts, nil
}
