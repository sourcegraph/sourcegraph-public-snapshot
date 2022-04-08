package gqltestutil

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func (c *Client) GetInsights() ([]string, error) {
	const query = `
		query InsightViews {
			insightViews {
				nodes { id }
			}
		}
	`
	var resp struct {
		Data struct {
			InsightViews struct {
				Nodes []struct {
					Id string `json:"id"`
				} `json:"nodes"`
			} `json:"insightviews"`
		} `json:"data"`
	}
	err := c.GraphQL("", query, map[string]interface{}{}, &resp)
	if err != nil {
		return nil, errors.Wrap(err, "request GraphQL")
	}

	ids := []string{}
	for _, node := range resp.Data.InsightViews.Nodes {
		ids = append(ids, node.Id)
	}

	return ids, nil
}
