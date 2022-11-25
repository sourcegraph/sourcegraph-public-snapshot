package gitlab

import (
	"context"
	"time"
)

// GetVersion retrieves the version of the GitLab instance.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	if MockGetVersion != nil {
		return MockGetVersion(ctx)
	}
	time.Sleep(c.rateLimitMonitor.RecommendedWaitForBackgroundOp(1))

	resp, err := version(ctx, c.gqlClient)
	return resp.Metadata.Version, err
}
