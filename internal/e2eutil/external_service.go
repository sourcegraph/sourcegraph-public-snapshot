package e2eutil

import (
	"github.com/pkg/errors"
)

type AddExternalServiceInput struct {
	Kind        string `json:"kind"`
	DisplayName string `json:"displayName"`
	Config      string `json:"config"`
}

// TODO
func (c *Client) AddExternalService(input AddExternalServiceInput) error {
	const query = `
	mutation addExternalService($input: AddExternalServiceInput!) {
		addExternalService(input: $input) {
			id
		}
	}
`
	variables := map[string]interface{}{
		"input": input,
	}
	err := c.GraphQL("", query, variables, nil)
	if err != nil {
		return errors.Wrap(err, "request GraphQL")
	}

	return nil
}
