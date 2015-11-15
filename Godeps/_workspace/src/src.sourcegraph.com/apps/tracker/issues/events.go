package issues

import "time"

// Event represents an event that occurred around an issue.
type Event struct {
	Actor     User
	CreatedAt time.Time
	Type      EventType
	Rename    *Rename
}

type EventType string

const (
	Reopened EventType = "reopened"
	Closed   EventType = "closed"
	Renamed  EventType = "renamed"
)

func (et EventType) Valid() bool {
	switch et {
	case Reopened, Closed, Renamed:
		return true
	default:
		return false
	}
}

type Rename struct {
	From string
	To   string
}
