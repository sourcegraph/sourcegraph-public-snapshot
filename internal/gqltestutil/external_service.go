package gqltestutil

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AddExternalServiceInput struct {
	Kind        string `json:"kind"`
	DisplayName string `json:"displayName"`
	Config      string `json:"config"`
}

// AddExternalService adds a new external service with given input.
// It returns GraphQL node ID of newly created external service.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) AddExternalService(input AddExternalServiceInput) (string, error) {
	const query = `
mutation AddExternalService($input: AddExternalServiceInput!) {
	addExternalService(input: $input) {
		id
		warning
	}
}
`
	variables := map[string]any{
		"input": input,
	}
	var resp struct {
		Data struct {
			AddExternalService struct {
				ID      string `json:"id"`
				Warning string `json:"warning"`
			} `json:"addExternalService"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	// Return the ID along with the warning so we can still clean up properly.
	if resp.Data.AddExternalService.Warning != "" {
		return resp.Data.AddExternalService.ID, errors.New(resp.Data.AddExternalService.Warning)
	}
	return resp.Data.AddExternalService.ID, nil
}

type UpdateExternalServiceInput struct {
	ID          string  `json:"id"`
	DisplayName *string `json:"displayName"`
	Config      *string `json:"config"`
}

// UpdateExternalService updates existing external service with given input.
// It returns GraphQL node ID of updated external service.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) UpdateExternalService(input UpdateExternalServiceInput) (string, error) {
	const query = `
mutation UpdateExternalService($input: UpdateExternalServiceInput!) {
	updateExternalService(input: $input) {
		id
		warning
	}
}
`
	variables := map[string]any{
		"input": input,
	}
	var resp struct {
		Data struct {
			UpdateExternalService struct {
				ID      string `json:"id"`
				Warning string `json:"warning"`
			} `json:"updateExternalService"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return "", errors.Wrap(err, "request GraphQL")
	}

	// Return the ID along with the warning, so we can still clean up properly.
	if resp.Data.UpdateExternalService.Warning != "" {
		return resp.Data.UpdateExternalService.ID, errors.New(resp.Data.UpdateExternalService.Warning)
	}
	return resp.Data.UpdateExternalService.ID, nil
}

// CheckExternalService checks whether the external service exists.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) CheckExternalService(id string) error {
	const query = `
query CheckExternalService($id: ID!) {
	node(id: $id) {
		... on ExternalService {
			id
		}
	}
}
`
	variables := map[string]any{"id": id}
	var resp struct {
		Data struct {
			Node struct {
				ID string `json:"id"`
			} `json:"node"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}

// DeleteExternalService deletes the external service by given GraphQL node ID.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) DeleteExternalService(id string, async bool) error {
	const query = `
mutation DeleteExternalService($externalService: ID!, $async: Boolean!) {
	deleteExternalService(externalService: $externalService, async: $async) {
		alwaysNil
	}
}
`
	variables := map[string]any{"externalService": id, "async": async}

	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}
	return nil
}
