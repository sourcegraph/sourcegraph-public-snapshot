package events

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type EventName string

const (
	EventNameUnauthorized       EventName = "Unauthorized"
	EventNameAccessDenied       EventName = "AccessDenied"
	EventNameRateLimited        EventName = "RateLimited"
	EventNameCompletionsRequest EventName = "CompletionsRequest"
)

// Logger is an event logger.
type Logger interface {
	// LogEvent logs an event.
	LogEvent(event Event) error
}

// bigQueryLogger is a BigQuery event logger.
type bigQueryLogger struct {
	tableInserter *bigquery.Inserter
}

// NewBigQueryLogger returns a new BigQuery event logger.
func NewBigQueryLogger(projectID, dataset, table string) (Logger, error) {
	client, err := bigquery.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}
	return &bigQueryLogger{
		tableInserter: client.Dataset(dataset).Table(table).Inserter(),
	}, nil
}

// Event contains information to be logged.
type Event struct {
	Name           EventName
	SubscriptionID string
	Metadata       map[string]any
}

var _ bigquery.ValueSaver = bigQueryEvent{}

type bigQueryEvent struct {
	Name           string
	SubscriptionID string
	Metadata       json.RawMessage
	CreatedAt      time.Time
}

func (e bigQueryEvent) Save() (map[string]bigquery.Value, string, error) {
	values := map[string]bigquery.Value{
		"name":            e.Name,
		"subscription_id": e.SubscriptionID,
		"created_at":      e.CreatedAt,
	}
	if e.Metadata != nil {
		values["metadata"] = string(e.Metadata)
	}
	return values, "", nil
}

// LogEvent logs an event to BigQuery.
func (l *bigQueryLogger) LogEvent(event Event) error {
	if event.Name == "" {
		return errors.New("missing event name")
	} else if event.SubscriptionID == "" {
		event.SubscriptionID = "anonymous"
	}

	var metadata json.RawMessage
	if event.Metadata != nil {
		var err error
		metadata, err = json.Marshal(event.Metadata)
		if err != nil {
			return errors.Wrap(err, "marshaling metadata")
		}
	}

	err := l.tableInserter.Put(
		// NOTE: Using context.Background() because we still want to log the event in the
		// case of a request cancellation.
		context.Background(),
		bigQueryEvent{
			Name:           string(event.Name),
			SubscriptionID: event.SubscriptionID,
			Metadata:       metadata,
			CreatedAt:      time.Now(),
		},
	)
	if err != nil {
		return errors.Wrap(err, "inserting event")
	}
	return nil
}

type noopLogger struct{}

// NewNoopLogger returns a new no-op event logger.
func NewNoopLogger() Logger { return noopLogger{} }

func (noopLogger) LogEvent(event Event) error { return nil }
