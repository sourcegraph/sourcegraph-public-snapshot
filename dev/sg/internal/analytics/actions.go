package analytics

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/sourcegraph/sourcegraph/dev/okay"
)

// Submit pushes all persisted events to OkayHQ.
func Submit(okayToken string, gitHubLogin string) error {
	events, err := Load()
	if err != nil {
		return err
	}

	client := okay.NewClient(http.DefaultClient, okayToken)
	for _, ev := range events {
		// discard everything but the latest version of events. if event versions are
		// migrate-able, do the migrations here.
		if ev.Properties["event_version"] != eventVersion {
			continue
		}

		// clean up data
		ev.Labels = append(ev.Labels, "sg-analytics")
		for k, v := range ev.Properties {
			if len(v) == 0 {
				delete(ev.Properties, k)
			}
		}

		// push to okayhq
		if err := client.Push(ev); err != nil {
			return err
		}
	}

	return client.Flush()
}

// Persist stores all events in context to disk.
func Persist(ctx context.Context, command string, flags []string) error {
	store := getStore(ctx)
	if store == nil {
		return nil
	}
	return store.Persist(command, flags)
}

// Reset deletes all persisted events.
func Reset() error {
	p, err := eventsPath()
	if err != nil {
		return err
	}
	return os.Remove(p)
}

// Load retrieves all persisted events.
func Load() ([]*okay.Event, error) {
	p, err := eventsPath()
	if err != nil {
		return nil, err
	}

	file, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var events []*okay.Event
	for scanner.Scan() {
		// Don't worry too much about malformed events, analytics are relatively optional
		// so just grab what we can.
		var event okay.Event
		if err := json.Unmarshal(scanner.Bytes(), &event); err == nil {
			events = append(events, &event)
		}
	}
	return events, nil
}
