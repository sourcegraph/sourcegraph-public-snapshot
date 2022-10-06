package database

import (
	"context"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type WebhookStore interface {
	basestore.ShareableStore

	Create(ctx context.Context, newWebhook *types.Webhook) (*types.Webhook, error)
	Delete(ctx context.Context, id string) error
	Update(ctx context.Context, newWebhook *types.Webhook) (*types.Webhook, error)
	List(ctx context.Context) ([]*types.Webhook, error)
}

type webhookStore struct {
	*basestore.Store

	logger log.Logger
	key    encryption.Key
}

var _ WebhookStore = &webhookStore{}

// Create the webhook
func (s *webhookStore) Create(ctx context.Context, newWebhook *types.Webhook) (*types.Webhook, error) {
	// TODO(sashaostrikov) implement this method
	panic("implement this method")
}

// Delete the webhook
func (s *webhookStore) Delete(ctx context.Context, id string) error {
	// TODO(sashaostrikov) implement this method
	panic("implement this method")
}

// Update the webhook
func (s *webhookStore) Update(ctx context.Context, newWebhook *types.Webhook) (*types.Webhook, error) {
	// TODO(sashaostrikov) implement this method
	panic("implement this method")
}

// List the webhooks
func (s *webhookStore) List(ctx context.Context) ([]*types.Webhook, error) {
	// TODO(sashaostrikov) implement this method
	panic("implement this method")
}
