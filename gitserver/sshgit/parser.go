package sshgit

import (
	"reflect"

	"github.com/AaronO/go-git-http"
)

// emptyCommitID is used to signify that a branch was created or deleted.
const emptyCommitID = "0000000000000000000000000000000000000000"

// collapseDuplicateEvents transforms a githttp event list such that adjacent
// equivalent events are collapsed into a single event.
func collapseDuplicateEvents(eventsDup []githttp.Event) []githttp.Event {
	events := []githttp.Event{}
	var previousEvent githttp.Event
	for _, e := range eventsDup {
		if !reflect.DeepEqual(e, previousEvent) {
			events = append(events, e)
		}
		previousEvent = e
	}
	return events
}
