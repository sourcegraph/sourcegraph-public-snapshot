package types

import (
	"net/http"
	"time"
)

type WebhookLog struct {
	ID                int64
	ReceivedAt        time.Time
	ExternalServiceID *int64
	StatusCode        int
	Request           WebhookLogMessage
	Response          WebhookLogMessage
}

type WebhookLogMessage struct {
	Header  http.Header
	Body    []byte
	Method  string `json:",omitempty"`
	URL     string `json:",omitempty"`
	Version string `json:",omitempty"`
}
