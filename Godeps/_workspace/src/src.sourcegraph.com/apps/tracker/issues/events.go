package issues

import "time"

// Event represents an event that occurred around an issue.
type Event struct {
	ID        uint64
	Actor     User
	CreatedAt time.Time
	Type      EventType
	Rename    *Rename
}

// EventType is the type of an event.
type EventType string

const (
	// Reopened is when an issue is reopened.
	Reopened EventType = "reopened"
	// Closed is when an issue is closed.
	Closed EventType = "closed"
	// Renamed is when an issue is renamed.
	Renamed EventType = "renamed"
)

// Valid returns non-nil error if the event type is invalid.
func (et EventType) Valid() bool {
	switch et {
	case Reopened, Closed, Renamed:
		return true
	default:
		return false
	}
}

// Rename provides details for a Renamed event.
type Rename struct {
	From string
	To   string
}
