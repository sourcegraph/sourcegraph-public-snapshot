package papertrail

import (
	"net/http"
	"time"
)

// SearchOptions specify parameters to the Papertrail search HTTP API endpoint.
type SearchOptions struct {
	// SystemID filters the results to only those from the specified system.
	SystemID string `url:"system_id,omitempty"`

	// GroupID filters the results to only those from the specified group.
	GroupID string `url:"group_id,omitempty"`

	// Query filters the results to only those containing the query string.
	Query string `url:"q,omitempty"`

	// MaxID specifies that the response should only contain events whose ID is
	// less than the specified ID. This is used when viewing older events.
	MaxID string `url:"max_id,omitempty"`

	// MinID specifies that the response should only contain events whose ID is
	// greater than the specified ID.
	MinID string `url:"min_id,omitempty"`

	// MaxTime specifies that the response should only contain events before
	// this time.
	MaxTime time.Time `url:"max_time,unix,omitempty"`

	// MinTime specifies that the response should only contain events after this
	// time.
	MinTime time.Time `url:"min_time,unix,omitempty"`
}

// A SearchResponse is the response from the Papertrail HTTP API's search
// endpoint.
type SearchResponse struct {
	// MinID is the smallest event ID presented.
	MinID string `json:"min_id"`

	// MaxID is the highest event ID presented.
	MaxID string `json:"max_id"`

	// ReachedTimeLimit is whether Papertrail's per-request time lmit was
	// reached before a full set of events was found.
	ReachedTimeLimit bool `json:"reached_time_limit"`

	// ReachedBeginning means that the entire searchable duration has been
	// examined and no more matching messages are available.
	ReachedBeginning bool `json:"reached_beginning"`

	// MinTimeAt is the earliest event time presented.
	MinTimeAt time.Time `json:"min_time_at"`

	// Events holds the log messages in the response.
	Events []*Event `json:"events"`
}

// An Event is a log entry.
type Event struct {
	// ID is the unique Papertrail message ID (a 64-bit integer in base-10
	// string format).
	ID string `json:"id"`

	// ReceivedAt is when Papertrail received the log entry (in ISO 8601
	// timestamp format in the original JSON).
	ReceivedAt time.Time `json:"received_at"`

	// DisplayReceivedAt is a human-readable string of ReceivedAt in the
	// timezone of the API token owner.
	DisplayReceivedAt string `json:"display_received_at"`

	// SourceName is the sender name.
	SourceName string `json:"source_name"`

	// SourceID is the unique Papertrail sender ID.
	SourceID int `json:"source_id"`

	// SourceIP is the IP address that originated this log entry.
	SourceIP string `json:"source_ip"`

	// Facility is the syslog facility.
	Facility string `json:"facility"`

	// Severity is the syslog severity.
	Severity string `json:"severity"`

	// Program is the syslog "tag" or "program" field if set. If not set, it is
	// nil (JSON null).
	Program *string `json:"program"`

	// Message is the log entry's message.
	Message string `json:"message"`
}

// Search returns log events that match the parameters specified in opt.
func (c *Client) Search(opt SearchOptions) (*SearchResponse, *http.Response, error) {
	if !opt.MaxTime.IsZero() {
		opt.MaxTime = opt.MaxTime.In(time.UTC)
	}
	if !opt.MinTime.IsZero() {
		opt.MinTime = opt.MinTime.In(time.UTC)
	}

	req, err := c.NewRequest("GET", "events/search.json", opt, nil)
	if err != nil {
		return nil, nil, err
	}

	var searchResp *SearchResponse
	resp, err := c.Do(req, &searchResp)
	if err != nil {
		return nil, resp, err
	}

	return searchResp, resp, err
}
