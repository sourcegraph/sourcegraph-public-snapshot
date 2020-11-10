package httpapi

import (
	"context"
	"fmt"
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/repos"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDispatchSuccess(t *testing.T) {
	h := GithubWebhook{}
	var called bool
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called = true
		return nil
	})

	ctx := context.Background()
	require.NoError(t, h.Dispatch(ctx, "test-event-1", nil, nil))
	assert.True(t, called)
}

func TestDispatchNoHandler(t *testing.T) {
	h := GithubWebhook{}
	ctx := context.Background()
	// no op
	require.NoError(t, h.Dispatch(ctx, "test-event-1", nil, nil))
}

func TestDispatchSuccessMultiple(t *testing.T) {
	h := GithubWebhook{}
	var called int
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})

	ctx := context.Background()
	require.NoError(t, h.Dispatch(ctx, "test-event-1", nil, nil))
	assert.Equal(t, 2, called)
}

func TestDispatchError(t *testing.T) {
	h := GithubWebhook{}
	var called int
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return fmt.Errorf("oh dear")
	})
	h.Register("test-event-1", func(ctx context.Context, svc *repos.ExternalService, payload interface{}) error {
		called++
		return nil
	})

	ctx := context.Background()
	require.Error(t, h.Dispatch(ctx, "test-event-1", nil, nil))
	assert.Equal(t, 1, called)
}
