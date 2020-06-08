package e2eutil

import (
	"github.com/pkg/errors"
)

type AddExternalServiceInput struct {
	Kind        string `json:"kind"`
	DisplayName string `json:"displayName"`
	Config      string `json:"config"`
}

// AddExternalService adds a new external service with given input.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) AddExternalService(input AddExternalServiceInput) error {
	const query = `
mutation AddExternalService($input: AddExternalServiceInput!) {
	addExternalService(input: $input) {
		id
		warning
	}
}
`
	variables := map[string]interface{}{
		"input": input,
	}
	var resp struct {
		Data struct {
			AddExternalService struct {
				Warning string `json:"warning"`
			} `json:"addExternalService"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, variables, &resp)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}

	if resp.Data.AddExternalService.Warning != "" {
		return errors.New(resp.Data.AddExternalService.Warning)
	}
	return nil
}
