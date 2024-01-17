package attribution

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	service := NewService(observation.TestContextTB(t), gateway)
	attribution, err := service.SnippetAttribution(context.Background(), "snippet", 3)
	require.NoError(t, err)
	require.Equal(t, &SnippetAttributions{
		RepositoryNames: []string{"repo1", "repo2"},
		TotalCount:      2,
		LimitHit:        false,
	}, attribution)
}
