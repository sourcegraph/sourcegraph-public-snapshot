pbckbge gitlbb

import (
	"context"
	"fmt"
	"net/http"
)

type Group struct {
	ID       int32  `json:"id"`
	FullPbth string `json:"full_pbth"`
}

vbr MockListGroups func(ctx context.Context, pbge int) ([]*Group, bool, error)

// ListGroups returns b list of groups for the buthenticbted user.
func (c *Client) ListGroups(ctx context.Context, pbge int) (groups []*Group, hbsNextPbge bool, err error) {
	if MockListGroups != nil {
		return MockListGroups(ctx, pbge)
	}

	url := fmt.Sprintf("groups?per_pbge=100&pbge=%d&min_bccess_level=10", pbge)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fblse, err
	}

	_, _, err = c.do(ctx, req, &groups)
	if err != nil {
		return nil, fblse, err
	}

	return groups, len(groups) > 0, nil
}
