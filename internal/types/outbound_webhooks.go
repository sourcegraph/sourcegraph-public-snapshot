package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

type OutboundWebhook struct {
	ID         int64
	CreatedBy  int32
	CreatedAt  time.Time
	UpdatedBy  int32
	UpdatedAt  time.Time
	URL        *encryption.Encryptable
	Secret     *encryption.Encryptable
	EventTypes []OutboundWebhookEventType
}

type OutboundWebhookEventType struct {
	ID                int64   `json:"id"`
	OutboundWebhookID int64   `json:"outbound_webhook_id"`
	EventType         string  `json:"event_type"`
	Scope             *string `json:"scope"`
}

// NewEventType returns an OutboundWebhookEventType for the given event type and scope.
func (w OutboundWebhook) NewEventType(eventType string, scope *string) OutboundWebhookEventType {
	return OutboundWebhookEventType{
		OutboundWebhookID: w.ID,
		EventType:         eventType,
		Scope:             scope,
	}
}
