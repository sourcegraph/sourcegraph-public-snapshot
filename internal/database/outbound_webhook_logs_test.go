package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func setupOutboundWebhookTest(t *testing.T, ctx context.Context, db DB, key encryption.Key) (user *types.User, webhook *types.OutboundWebhook) {
	t.Helper()

	userStore := db.Users()
	user, err := userStore.Create(ctx, NewUser{Username: "test"})
	require.NoError(t, err)

	webhookStore := db.OutboundWebhooks(key)
	webhook = newTestWebhook(t, user, ScopedEventType{EventType: "foo"})
	err = webhookStore.Create(ctx, webhook)
	require.NoError(t, err)

	return
}
