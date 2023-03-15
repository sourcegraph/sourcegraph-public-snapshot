package gitlab

import (
	"context"
	"fmt"
	"net/http"
)

type Group struct {
	ID       int32  `json:"id"`
	FullPath string `json:"full_path"`
}

var MockListGroups func(ctx context.Context, page int) ([]*Group, bool, error)

// ListGroups returns a list of groups for the authenticated user.
func (c *Client) ListGroups(ctx context.Context, page int) (groups []*Group, hasNextPage bool, err error) {
	if MockListGroups != nil {
		return MockListGroups(ctx, page)
	}

	url := fmt.Sprintf("groups?per_page=100&page=%d&min_access_level=10", page)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, false, err
	}

	_, _, err = c.do(ctx, req, &groups)
	if err != nil {
		return nil, false, err
	}

	return groups, len(groups) > 0, nil
}
