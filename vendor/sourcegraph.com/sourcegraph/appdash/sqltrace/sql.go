// Package sqltrace implements utility types for tracing SQL queries.
package sqltrace

import (
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

// SQLEvent is an SQL query event for use with appdash. It's primary function
// is to measure the time between when the query is sent and later received.
type SQLEvent struct {
	SQL        string
	Tag        string
	ClientSend time.Time
	ClientRecv time.Time
}

// Schema implements the appdash Event interface by returning this event's
// constant schema string, "SQL".
func (SQLEvent) Schema() string { return "SQL" }

// Important implements the appdash ImportantEvent by returning the SQL and Tag
// keys.
func (SQLEvent) Important() []string { return []string{"SQL", "Tag"} }

// Start implements the appdash TimespanEvent interface by returning the time
// at which the SQL query was sent out.
func (e SQLEvent) Start() time.Time { return e.ClientSend }

// End implements the appdash TimespanEvent interface by returning the time at
// which the SQL query returned / was received.
func (e SQLEvent) End() time.Time { return e.ClientRecv }

func init() { appdash.RegisterEvent(SQLEvent{}) }
