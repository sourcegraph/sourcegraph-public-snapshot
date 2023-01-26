package gitlab

import (
	"context"
	"net/http"

	"github.com/peterhellberg/link"
)

type Group struct {
	ID       int32  `json:"id"`
	FullPath string `json:"full_path"`
}

var MockListGroups func(ctx context.Context, pageURL *string) ([]*Group, *string, error)

// ListGroups returns a list of groups for the authenticated user.
func (c *Client) ListGroups(ctx context.Context, pageURL *string) (groups []*Group, nextPageURL *string, err error) {
	if MockListGroups != nil {
		return MockListGroups(ctx, pageURL)
	}

	url := "groups?per_page=100&min_access_level=10"
	if pageURL != nil {
		url = *pageURL
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, nil, err
	}

	respHeader, _, err := c.do(ctx, req, &groups)
	if err != nil {
		return nil, nil, err
	}

	if l := link.Parse(respHeader.Get("Link"))["next"]; l != nil {
		nextPageURL = &l.URI
	}

	return groups, nextPageURL, nil
}
