package client

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

func (c *Client) Stats(ctx context.Context) (stats *lsif.LSIFStats, err error) {
	_, err = c.do(ctx, &lsifRequest{path: "/stats"}, &stats)
	return
}
