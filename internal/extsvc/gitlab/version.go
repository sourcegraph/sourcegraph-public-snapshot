package gitlab

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/graphql"
)

// GetVersion retrieves the version of the GitLab instance.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	if MockGetVersion != nil {
		return MockGetVersion(ctx)
	}
	time.Sleep(c.rateLimitMonitor.RecommendedWaitForBackgroundOp(1))

	resp, err := graphql.Version(ctx, c.gqlClient)
	return resp.Metadata.Version, err
}
