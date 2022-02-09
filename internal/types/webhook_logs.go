package types

import (
	"net/http"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

type WebhookLog struct {
	ID                int64
	ReceivedAt        time.Time
	ExternalServiceID *int64
	WebhookID         *int32
	StatusCode        int
	Request           *EncryptableWebhookLogMessage
	Response          *EncryptableWebhookLogMessage
}

type WebhookLogMessage struct {
	Header  http.Header
	Body    []byte
	Method  string `json:",omitempty"`
	URL     string `json:",omitempty"`
	Version string `json:",omitempty"`
}

type EncryptableWebhookLogMessage = encryption.JSONEncryptable[WebhookLogMessage]

func NewUnencryptedWebhookLogMessage(value WebhookLogMessage) *EncryptableWebhookLogMessage {
	message, _ := encryption.NewUnencryptedJSON(value)
	return message
}

func NewEncryptedWebhookLogMessage(cipher, keyID string, key encryption.Key) *EncryptableWebhookLogMessage {
	return encryption.NewEncryptedJSON[WebhookLogMessage](cipher, keyID, key)
}
