package gitlab

import (
	"context"
	"net/http"
)

type Group struct {
	ID       int32  `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	FullName string `json:"full_name"`
}

var MockListGroups func(ctx context.Context) ([]*Group, error)

func (c *Client) ListGroups(ctx context.Context) (groups []*Group, err error) {
	if MockListGroups != nil {
		return MockListGroups(ctx)
	}

	req, err := http.NewRequest("GET", "groups/", nil)
	if err != nil {
		return nil, err
	}

	if _, _, err := c.do(ctx, req, &groups); err != nil {
		return nil, err
	}

	return groups, nil
}
