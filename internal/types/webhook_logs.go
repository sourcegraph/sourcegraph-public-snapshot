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
	return NewUnencryptedWebhookLogMessageWithKey(value, nil)
}

func NewUnencryptedWebhookLogMessageWithKey(value WebhookLogMessage, key encryption.Key) *EncryptableWebhookLogMessage {
	message, _ := encryption.NewUnencryptedJSONWithKey(value, key)
	return message
}

func NewEncryptedWebhookLogMessage(cipher, keyID string, key encryption.Key) *EncryptableWebhookLogMessage {
	return encryption.NewEncryptedJSON[WebhookLogMessage](cipher, keyID, key)
}
