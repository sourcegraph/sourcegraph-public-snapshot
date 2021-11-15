package amplitude

import (
	"encoding/json"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

// EventPayload represents the structure of the payloads we send to Amplitude
// when we log an event. We can send more than one event in Events, but we currently
// send events in individual payloads.
type EventPayload struct {
	APIKey string           `json:"api_key"`
	Events []AmplitudeEvent `json:"events"`
}

// AmplitudeEvent represents the fields that make up an event
// in Amplitude.
type AmplitudeEvent struct {
	UserID          string          `json:"user_id"`
	DeviceID        string          `json:"device_id"`
	EventID         int32           `json:"event_id"`
	InsertID        string          `json:"insert_id"`
	EventType       string          `json:"event_type"`
	EventProperties json.RawMessage `json:"event_properties,omitempty"`
	UserProperties  json.RawMessage `json:"user_properties,omitempty"`
	Time            int64           `json:"time,omitempty"`
}

// UserProperties contains the list of user properties we collect and send to Amplitude.
type UserProperties struct {
	AnonymousUserID string              `json:"anonymous_user_id"`
	FeatureFlags    featureflag.FlagSet `json:"feature_flags"`
}
