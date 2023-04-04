package types

import (
	"time"

	"github.com/sourcegraph/sourcegraph/internal/encryption"
)

type OutboundWebhookLog struct {
	ID                int64
	JobID             int64
	OutboundWebhookID int64
	SentAt            time.Time
	StatusCode        int
	Request           *EncryptableWebhookLogMessage
	Response          *EncryptableWebhookLogMessage
	Error             *encryption.Encryptable
}

const OutboundWebhookLogUnsentStatusCode int = 0
