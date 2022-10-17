package okay

import "time"

// customEvent represents a custom event sent to the OkayHQ API.
type customEvent struct {
	// Event is the event type, should be always set to "custom".
	Event string `json:"event"`
	// Name is the custom event name, used to select those events amonst others to build dashboards.
	Name string `json:"customEventName"`
	// Timestamp is the time at which this event occured.
	Timestamp time.Time `json:"timestamp"`
	// Identity ties this specific event to a particular user, enabling to filter events on various group predicates
	// such as teams, organizations, etc ... (optional).
	Identity *eventIdentity `json:"identity,omitempty"`
	// UniqueKey lists the property keys that are used to uniquely identify this event.
	//
	// Sending another event with the same UniqueKey results in overwritting the previous event,
	// enabling to replay events with historical data or to correct incorrect events that were previously sent.
	UniqueKey []string `json:"uniqueKey,omitempty"`
	// Metrics are a map of okayMetric whose keys are the metric name.
	Metrics map[string]Metric `json:"metrics"`
	// Properties are a map of additonal metadata (optional).
	Properties map[string]string `json:"properties,omitempty"`
	// Labels are a list of strings used to tag the event.
	Labels []string `json:"labels,omitempty"`
}

// eventIdentity represents the identity to attach to an event.
type eventIdentity struct {
	// Type represents from where this identity is registered, should always be "sourceControlLogin".
	Type string `json:"type"`
	// User is the unique identifier to reference this identity amongst its Type.
	User string `json:"user"`
}

// Event represents a custom Event to be sent to OkayHQ.
type Event struct {
	// Name is the custom event name, used to select those events amonst others to build dashboards.
	Name string
	// Timestamp is the time at which this event occured.
	Timestamp time.Time
	// Metrics are a map of okayMetric whose keys are the metric name.
	Metrics map[string]Metric
	// Labels are used to enable advanced filtering and group bys.
	Labels []string
	// UniqueKey lists the property keys that are used to uniquely identify this event.
	// This will allow versioning and replaying events to correct errors or add new metadata.
	//
	// Sending another event with the same UniqueKey results in overwritting the previous event,
	// enabling to replay events with historical data or to correct incorrect events that were previously sent.
	UniqueKey []string
	// Properties are a map of additonal metadata to enable filtering and grouping metrics.
	// Property names cannot have ".", please use "_" instead.
	Properties map[string]string

	// Optional fields below

	// GitHub login this event is attached to (optional).
	GitHubLogin string `json:",omitempty"`
	// OkayURL is used to generate a clickable link in OkayHQ's UI when browsing this
	// event (optional).
	//
	// Usually a link to get more information about the event from the system that generated it.
	OkayURL string `json:",omitempty"`
}
