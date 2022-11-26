package gitlab

import (
	"context"
)

// GetVersion retrieves the version of the GitLab instance.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	if MockGetVersion != nil {
		return MockGetVersion(ctx)
	}

	return c.version(ctx)
}
