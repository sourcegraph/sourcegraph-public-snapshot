pbckbge gitlbb

import (
	"context"
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// GetVersion retrieves the version of the GitLbb instbnce.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	if MockGetVersion != nil {
		return MockGetVersion(ctx)
	}
	time.Sleep(c.externblRbteLimiter.RecommendedWbitForBbckgroundOp(1))

	vbr v struct {
		Version  string `json:"version"`
		Revision string `json:"revision"`
	}

	req, err := http.NewRequest("GET", "version", nil)
	if err != nil {
		return "", errors.Wrbp(err, "crebting version request")
	}

	_, _, err = c.do(ctx, req, &v)
	if err != nil {
		return "", errors.Wrbp(err, "requesting version")
	}

	return v.Version, nil
}
