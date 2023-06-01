package events

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Logger is an event logger.
type Logger interface {
	// LogEvent logs an event. spanCtx should only be used to extract the span,
	// event logging should use a background.Context to avoid being cancelled
	// when a request ends.
	LogEvent(spanCtx context.Context, event Event) error
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
	return &instrumentedLogger{
		Scope: "bigQueryLogger",
		Logger: &bigQueryLogger{
			tableInserter: client.Dataset(dataset).Table(table).Inserter(),
		},
	}, nil
}

// Event contains information to be logged.
type Event struct {
	Name       codygateway.EventName
	Source     string
	Identifier string
	Metadata   map[string]any
}

var _ bigquery.ValueSaver = bigQueryEvent{}

type bigQueryEvent struct {
	Name       string
	Source     string
	Identifier string
	Metadata   json.RawMessage
	CreatedAt  time.Time
}

func (e bigQueryEvent) Save() (map[string]bigquery.Value, string, error) {
	values := map[string]bigquery.Value{
		"name":       e.Name,
		"source":     e.Source,
		"identifier": e.Identifier,
		"created_at": e.CreatedAt,
	}
	if e.Metadata != nil {
		values["metadata"] = string(e.Metadata)
	}
	return values, "", nil
}

// LogEvent logs an event to BigQuery.
func (l *bigQueryLogger) LogEvent(spanCtx context.Context, event Event) (err error) {
	if event.Name == "" || event.Source == "" || event.Identifier == "" {
		return errors.New("missing event name, source or identifier")
	}

	var metadata json.RawMessage
	if event.Metadata != nil {
		var err error
		metadata, err = json.Marshal(event.Metadata)
		if err != nil {
			return errors.Wrap(err, "marshaling metadata")
		}
	}

	if err := l.tableInserter.Put(
		backgroundContextWithSpan(spanCtx),
		bigQueryEvent{
			Name:       string(event.Name),
			Source:     event.Source,
			Identifier: event.Identifier,
			Metadata:   metadata,
			CreatedAt:  time.Now(),
		},
	); err != nil {
		return errors.Wrap(err, "inserting BigQuery event")
	}
	return nil
}

type stdoutLogger struct {
	logger log.Logger
}

// NewStdoutLogger returns a new stdout event logger.
func NewStdoutLogger(logger log.Logger) Logger {
	// Wrap in instrumentation - not terribly interesting traces, but useful to
	// demo tracing in dev.
	return &instrumentedLogger{
		Scope:  "stdoutLogger",
		Logger: &stdoutLogger{logger: logger.Scoped("events", "event logger")},
	}
}

func (l *stdoutLogger) LogEvent(spanCtx context.Context, event Event) error {
	trace.Logger(spanCtx, l.logger).Debug("LogEvent",
		log.Object("event",
			log.String("name", string(event.Name)),
			log.String("source", event.Source),
			log.String("identifier", event.Identifier),
		),
	)
	return nil
}
