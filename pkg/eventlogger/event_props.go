package eventlogger

import "encoding/json"

// Payload represents context about a user event to be tracked
type Payload struct {
	DeviceInfo *DeviceInfo `json:"device_info,omitempty"`
	Header     *Header     `json:"header,omitempty"`
	Events     []*Event    `json:"events,omitempty"`
	UserInfo   *UserInfo   `json:"user_info,omitempty"`
	BatchInfo  *BatchInfo  `json:"batch_info,omitempty"`
}

// Header represents environment-level properties
type Header struct {
	// TODO(sqs): It is intentional that the Go field name is SiteID and the JSON
	// field name is app_id (which is the name in our telemetry backend).
	SiteID string `json:"app_id,omitempty"`
	Env    string `json:"env,omitempty"`
}

// DeviceInfo represents platform- and device-level properties
type DeviceInfo struct {
	Platform         string `json:"platform,omitempty"`
	TrackerNamespace string `json:"tracker_namespace,omitempty"`
}

// UserInfo represents user-level properties
type UserInfo struct {
	DomainUserID string `json:"domain_user_id"`
	Email        string `json:"email,omitempty"`
}

// BatchInfo represents event group/batch-level properties
type BatchInfo struct {
	BatchID     string `json:"batch_id,omitempty"`
	TotalEvents int    `json:"total_events,omitempty"`
	ServerTime  string `json:"server_time,omitempty"`
}

// Event represents event-level properties
type Event struct {
	Type            string   `json:"type,omitempty"`
	Context         *Context `json:"ctx,omitempty"`
	EventID         string   `json:"event_id,omitempty"`
	ClientTimestamp int64    `json:"client_tstamp,omitempty"`
}

// Context represents custom event-level context/properties that can be passed with an event
type Context struct {
	EventLabel string          `json:"event_label"`
	Backend    json.RawMessage `json:"backend,omitempty"`
}
