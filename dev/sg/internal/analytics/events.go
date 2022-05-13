package analytics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/sourcegraph/sourcegraph/dev/okay"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

const eventVersion = "v0"

// eventStore tracks events for a single sg command run.
type eventStore struct {
	sgVersion string
	events    []*okay.Event
}

// Persist is called once per sg run. All in this run events are correlated with a single
// run ID.
func (s *eventStore) Persist(command string, flagsUsed []string) error {
	runID := uuid.NewString()

	// Finalize events
	for _, ev := range s.events {
		// Create additional identifying keys
		ev.UniqueKey = append(ev.UniqueKey,
			"context",
			"event_name",
			"event_version",
			"run_id")

		// Identifying keys
		ev.Properties["context"] = "sg"
		ev.Properties["event_name"] = ev.Name
		ev.Properties["event_version"] = eventVersion
		ev.Properties["run_id"] = runID

		// Context
		ev.Properties["command"] = command
		ev.Properties["sg_version"] = s.sgVersion
		if len(flagsUsed) > 0 {
			ev.Properties["flags_used"] = strings.Join(flagsUsed, ",")
		}
	}

	// Persist events to disk
	return storeEvents(s.events)
}

func eventsPath() (string, error) {
	home, err := root.GetSGHomePath()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "events"), nil
}

func storeEvents(events []*okay.Event) error {
	p, err := eventsPath()
	if err != nil {
		return err
	}

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(p, os.O_APPEND|os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return err
	}

	// Generate newline-separated representation of events
	for _, ev := range events {
		b, err := json.Marshal(ev)
		if err != nil {
			return err
		}
		f.Write(b)
		f.WriteString("\n")
	}

	return nil
}
