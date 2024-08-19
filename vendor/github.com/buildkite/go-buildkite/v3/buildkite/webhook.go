package buildkite

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"hash"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	// EventTypeHeader is the Buildkite header key used to pass the event type
	EventTypeHeader = "X-Buildkite-Event"
	// SignatureHeader is the Buildkite header key used to pass the HMAC hexdigest.
	SignatureHeader = "X-Buildkite-Signature"
)

var (
	// eventTypeMapping maps webhook types to their corresponding Buildkite structs
	eventTypeMapping = map[string]string{
		"agent.connected":    "AgentConnectedEvent",
		"agent.disconnected": "AgentDisconnectedEvent",
		"agent.lost":         "AgentLostEvent",
		"agent.stopped":      "AgentStoppedEvent",
		"agent.stopping":     "AgentStoppingEvent",
		"build.failing":      "BuildFailingEvent",
		"build.finished":     "BuildFinishedEvent",
		"build.running":      "BuildRunningEvent",
		"build.scheduled":    "BuildScheduledEvent",
		"job.activated":      "JobActivatedEvent",
		"job.finished":       "JobFinishedEvent",
		"job.scheduled":      "JobScheduledEvent",
		"job.started":        "JobStartedEvent",
		"ping":               "PingEvent",
	}
)

// WebHookType returns the event type of webhook request r.
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks
func WebHookType(r *http.Request) string {
	return r.Header.Get(EventTypeHeader)
}

// ParseWebHook parses the event payload. For recognized event types, a
// value of the corresponding struct type will be returned (as returned
// by Event.ParsePayload()). An error will be returned for unrecognized event
// types.
func ParseWebHook(messageType string, payload []byte) (interface{}, error) {
	eventType, ok := eventTypeMapping[messageType]
	if !ok {
		return nil, fmt.Errorf("unknown X-Buildkite-Event in message: %v", messageType)
	}

	event := Event{
		Type:       &eventType,
		RawPayload: (*json.RawMessage)(&payload),
	}
	return event.ParsePayload()
}

// genMAC generates the HMAC signature for a message provided the secret key
// and hashFunc.
func genMAC(message, key []byte, hashFunc func() hash.Hash) []byte {
	mac := hmac.New(hashFunc, key)
	mac.Write(message)
	return mac.Sum(nil)
}

// checkMAC reports whether messageMAC is a valid HMAC tag for message.
func checkMAC(message, messageMAC, key []byte, hashFunc func() hash.Hash) bool {
	expectedMAC := genMAC(message, key, hashFunc)
	return hmac.Equal(messageMAC, expectedMAC)
}

// validateSignature validates the signature for the given payload.
// signature is the Buildkite hash signature delivered in the X-Buildkite-Signature header.
// payload is the JSON payload sent by Buildkite Webhook.
// secretKey is the Buildkite Webhook token.
//
// Buildkite API docs: https://buildkite.com/docs/apis/webhooks#webhook-signature
func validateSignature(signature string, payload, secretKey []byte) error {
	timestamp, sig, err := getTimestampAndSignature(signature)
	if err != nil {
		return err
	}

	macPayload := fmt.Sprintf("%s.%s", timestamp, payload)

	if !checkMAC([]byte(macPayload), sig, secretKey, sha256.New) {
		return fmt.Errorf("payload signature check failed")
	}

	return nil
}

// getTimestampAndSignature splits the signature header into the timestamp and signature
// components.
// sig is the Buildkite hash signature value
func getTimestampAndSignature(sig string) (timestamp string, signature []byte, err error) {
	sigParts := strings.Split(sig, ",")
	if len(sigParts) != 2 {
		return "", nil, fmt.Errorf("X-Buildkite-Signature format is incorrect.")
	}

	ts, sg := sigParts[0], sigParts[1]

	timestamp = strings.Split(ts, "=")[1]
	sigStr := strings.Split(sg, "=")[1]
	signature, err = hex.DecodeString(sigStr)
	if err != nil {
		return "", nil, fmt.Errorf("error decoding signature %q: %v", sigStr, err)
	}

	return timestamp, signature, nil
}

// ValidatePayload validates an incoming Buildkite Webhook event request
// and returns the (JSON) payload.
// secretKey is the Buildkite Webhook token.
//
// Example usage:
func ValidatePayload(r *http.Request, secretKey []byte) (payload []byte, err error) {
	if payload, err = ioutil.ReadAll(r.Body); err != nil {
		return nil, err
	}

	sig := r.Header.Get(SignatureHeader)
	if sig == "" {
		return nil, fmt.Errorf("No %s header present on request", SignatureHeader)
	}

	if err = validateSignature(sig, payload, secretKey); err != nil {
		return nil, err
	}

	return payload, nil
}

// Event represents a Buildkite webhook event
type Event struct {
	Type       *string          `json:"type"`
	RawPayload *json.RawMessage `json:"payload,omitempty"`
}

// func (e Event) String() string {
// 	return Stringify(e)
// }

// ParsePayload parses the event payload. For recognized event types,
// a value of the corresponding struct type will be returned.
// An error will be returned for unrecognized event types.
//
// Example usage:
func (e *Event) ParsePayload() (payload interface{}, err error) {
	switch *e.Type {
	case "AgentConnectedEvent":
		payload = &AgentConnectedEvent{}
	case "AgentDisconnectedEvent":
		payload = &AgentDisconnectedEvent{}
	case "AgentLostEvent":
		payload = &AgentLostEvent{}
	case "AgentStoppedEvent":
		payload = &AgentStoppedEvent{}
	case "AgentStoppingEvent":
		payload = &AgentStoppingEvent{}
	case "BuildFailingEvent":
		payload = &BuildFailingEvent{}
	case "BuildFinishedEvent":
		payload = &BuildFinishedEvent{}
	case "BuildRunningEvent":
		payload = &BuildRunningEvent{}
	case "BuildScheduledEvent":
		payload = &BuildScheduledEvent{}
	case "JobActivatedEvent":
		payload = &JobActivatedEvent{}
	case "JobFinishedEvent":
		payload = &JobFinishedEvent{}
	case "JobScheduledEvent":
		payload = &JobScheduledEvent{}
	case "JobStartedEvent":
		payload = &JobStartedEvent{}
	case "PingEvent":
		payload = &PingEvent{}
	}
	err = json.Unmarshal(*e.RawPayload, &payload)
	return payload, err
}
