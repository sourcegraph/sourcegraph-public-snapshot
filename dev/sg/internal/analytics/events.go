package analytics

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/sourcegraph/sourcegraph/dev/okay"
	"github.com/sourcegraph/sourcegraph/dev/sg/root"
)

// eventStore tracks events for a single sg command run.
type eventStore struct {
	processor trace.SpanProcessor
}

// Persist is called once per sg run, at the end, to save events
func (s *eventStore) Persist(ctx context.Context) error {
	return s.processor.Shutdown(ctx)
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
