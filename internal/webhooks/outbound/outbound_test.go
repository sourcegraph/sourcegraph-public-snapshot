package outbound

import (
	"context"
	"testing"

	mockassert "github.com/derision-test/go-mockgen/testutil/assert"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestEnqueueWebhook(t *testing.T) {
	ctx := context.Background()
	payload := []byte(`"TEST"`)

	t.Run("store error", func(t *testing.T) {
		want := errors.New("mock error")
		store := database.NewMockOutboundWebhookJobStore()
		store.CreateFunc.SetDefaultReturn(nil, want)
		svc := &outboundWebhookService{store}

		have := svc.Enqueue(ctx, "type", nil, payload)
		assert.ErrorIs(t, have, want)
		mockassert.CalledOnce(t, store.CreateFunc)
	})

	t.Run("success", func(t *testing.T) {
		store := database.NewMockOutboundWebhookJobStore()
		store.CreateFunc.SetDefaultReturn(&types.OutboundWebhookJob{}, nil)
		svc := &outboundWebhookService{store}

		err := svc.Enqueue(ctx, "type", nil, payload)
		assert.NoError(t, err)
		mockassert.CalledOnce(t, store.CreateFunc)
	})
}
