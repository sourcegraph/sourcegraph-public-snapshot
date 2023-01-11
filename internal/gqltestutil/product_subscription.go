package gqltestutil

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ProductSubscription struct {
	License *struct {
		ProductNameWithBrand string `json:"productNameWithBrand"`
	} `json:"license"`
}

// ProductSubscription returns information of the current product subscription.
//
// This method requires the authenticated user to be a site admin.
func (c *Client) ProductSubscription() (*ProductSubscription, error) {
	const query = `
query ProductSubscription {
	site {
		productSubscription {
			license {
				productNameWithBrand
			}
		}
	}
}
`

	var resp struct {
		Data struct {
			Site struct {
				ProductSubscription ProductSubscription `json:"productSubscription"`
			} `json:"site"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, nil, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}
	return &resp.Data.Site.ProductSubscription, nil
}
