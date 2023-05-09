package events

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Logger is a BigQuery event logger.
// todo make this an interface and add noopLogger
type Logger struct {
	client  *bigquery.Client
	dataset string
	table   string
}

// NewLogger returns a new BigQuery event logger.
func NewLogger(projectID, dataset, table string) (*Logger, error) {
	client, err := bigquery.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, errors.Wrap(err, "creating BigQuery client")
	}
	return &Logger{
		client:  client,
		dataset: dataset,
		table:   table,
	}, nil
}

type EventName string

const (
	EventNameUnauthorized       EventName = "Unauthorized"
	EventNameAccessDenied       EventName = "AccessDenied"
	EventNameRateLimited        EventName = "RateLimited"
	EventNameCompletionsRequest EventName = "CompletionsRequest"
)

// Event contains information to be sent to BigQuery.
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
	return map[string]bigquery.Value{
		"name":            e.Name,
		"subscription_id": e.SubscriptionID,
		"metadata":        e.Metadata,
		"created_at":      e.CreatedAt,
	}, "", nil
}

// LogEvent logs an event to BigQuery.
func (l *Logger) LogEvent(event Event) error {
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

	err := l.client.Dataset(l.dataset).Table(l.table).Inserter().Put(
		// NOTE: Using context.Background() because we still want to log the event in the
		// case of a request cancellation.
		context.Background(),
		bigQueryEvent{
			Name:           string(event.Name),
			SubscriptionID: event.SubscriptionID,
			Metadata:       metadata,
			CreatedAt:      time.Now(),
		},
	) // todo l.tableInserter?
	if err != nil {
		return errors.Wrap(err, "inserting event")
	}
	return nil
}
