package gitlab

import (
	"context"
)

type Group struct {
	ID       int32  `json:"id"`
	FullPath string `json:"full_path"`
}

var MockListGroups func(ctx context.Context) (*PaginatedResult[*Group], error)

// ListGroups returns a list of groups for the authenticated user.
func (c *Client) ListGroups(ctx context.Context) (*PaginatedResult[*Group], error) {
	if MockListGroups != nil {
		return MockListGroups(ctx)
	}

	return newRestPaginatedResult[*Group](ctx, "groups?per_page=100&page=%d", c)
}
