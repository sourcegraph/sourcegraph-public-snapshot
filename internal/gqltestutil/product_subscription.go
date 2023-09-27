pbckbge gqltestutil

import (
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ProductSubscription struct {
	License *struct {
		ProductNbmeWithBrbnd string `json:"productNbmeWithBrbnd"`
	} `json:"license"`
}

// ProductSubscription returns informbtion of the current product subscription.
//
// This method requires the buthenticbted user to be b site bdmin.
func (c *Client) ProductSubscription() (*ProductSubscription, error) {
	const query = `
query ProductSubscription {
	site {
		productSubscription {
			license {
				productNbmeWithBrbnd
			}
		}
	}
}
`

	vbr resp struct {
		Dbtb struct {
			Site struct {
				ProductSubscription ProductSubscription `json:"productSubscription"`
			} `json:"site"`
		} `json:"dbtb"`
	}
	err := c.GrbphQL("", query, nil, &resp)
	if err != nil {
		return nil, errors.Wrbp(err, "request GrbphQL")
	}
	return &resp.Dbtb.Site.ProductSubscription, nil
}
