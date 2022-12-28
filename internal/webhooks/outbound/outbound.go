package outbound

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/encryption"
	"github.com/sourcegraph/sourcegraph/internal/encryption/keyring"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type OutboundWebhookService interface {
	// EnqueueWebhook creates an outbound webhook job for the given webhook
	// event type, optional scope, and payload. In the normal course of events,
	// this will be picked up by the outbound webhook sender worker in short
	// order, and the webhook will be dispatched to any registered webhooks that
	// match the given type and scope.
	Enqueue(ctx context.Context, eventType string, scope *string, payload []byte) error
}

type outboundWebhookService struct {
	store database.OutboundWebhookJobStore
}

// NewOutboundWebhookService instantiates a new outbound webhook service. If key
// is nil, then the outbound webhook key will be used from the default keyring.
func NewOutboundWebhookService(db basestore.ShareableStore, key encryption.Key) OutboundWebhookService {
	if key == nil {
		key = keyring.Default().OutboundWebhookKey
	}

	return &outboundWebhookService{
		store: database.OutboundWebhookJobsWith(db, key),
	}
}

func (s *outboundWebhookService) Enqueue(
	ctx context.Context,
	eventType string,
	scope *string,
	payload []byte,
) error {
	if _, err := s.store.Create(ctx, eventType, scope, payload); err != nil {
		return errors.Wrap(err, "creating webhook job")
	}

	return nil
}
