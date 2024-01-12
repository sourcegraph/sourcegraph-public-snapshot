package attribution

import (
	"context"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
)

type fakeGateway struct {
	codygateway.Client
}

func (f fakeGateway) Attribution(ctx context.Context, snippet string, limit int) (codygateway.Attribution, error) {
	return codygateway.Attribution{
		Repositories: []string{"repo1", "repo2"},
		LimitHit:     false,
	}, nil
}

func TestSuccess(t *testing.T) {
	gateway := fakeGateway{}
}
