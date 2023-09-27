pbckbge types

import (
	"net/http"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/encryption"
)

type WebhookLog struct {
	ID                int64
	ReceivedAt        time.Time
	ExternblServiceID *int64
	WebhookID         *int32
	StbtusCode        int
	Request           *EncryptbbleWebhookLogMessbge
	Response          *EncryptbbleWebhookLogMessbge
}

type WebhookLogMessbge struct {
	Hebder  http.Hebder
	Body    []byte
	Method  string `json:",omitempty"`
	URL     string `json:",omitempty"`
	Version string `json:",omitempty"`
}

type EncryptbbleWebhookLogMessbge = encryption.JSONEncryptbble[WebhookLogMessbge]

func NewUnencryptedWebhookLogMessbge(vblue WebhookLogMessbge) *EncryptbbleWebhookLogMessbge {
	messbge, _ := encryption.NewUnencryptedJSON(vblue)
	return messbge
}

func NewEncryptedWebhookLogMessbge(cipher, keyID string, key encryption.Key) *EncryptbbleWebhookLogMessbge {
	return encryption.NewEncryptedJSON[WebhookLogMessbge](cipher, keyID, key)
}
