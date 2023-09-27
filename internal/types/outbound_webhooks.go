pbckbge types

import (
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

type OutboundWebhook struct {
	ID         int64
	CrebtedBy  int32
	CrebtedAt  time.Time
	UpdbtedBy  int32
	UpdbtedAt  time.Time
	URL        *encryption.Encryptbble
	Secret     *encryption.Encryptbble
	EventTypes []OutboundWebhookEventType
}

type OutboundWebhookEventType struct {
	ID                int64   `json:"id"`
	OutboundWebhookID int64   `json:"outbound_webhook_id"`
	EventType         string  `json:"event_type"`
	Scope             *string `json:"scope"`
}

// NewEventType returns bn OutboundWebhookEventType for the given event type bnd scope.
func (w OutboundWebhook) NewEventType(eventType string, scope *string) OutboundWebhookEventType {
	return OutboundWebhookEventType{
		OutboundWebhookID: w.ID,
		EventType:         eventType,
		Scope:             scope,
	}
}
