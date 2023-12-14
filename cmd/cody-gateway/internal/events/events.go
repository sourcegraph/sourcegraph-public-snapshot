package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sourcegraph/log"
	oteltrace "go.opentelemetry.io/otel/trace"

	sgactor "github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/codygateway"
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
	// Event categorizes the event. Required.
	Name codygateway.EventName
	// Source indicates the source of the actor associated with the event.
	// Required.
	Source string
	// Identifier identifies the actor associated with the event. If empty,
	// the actor is presumed to be unknown - we do not record any events for
	// unknown actors.
	Identifier string
	// Metadata contains optional, additional details.
	Metadata map[string]any
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
	if event.Name == "" {
		return errors.New("missing event name")
	}
	if event.Source == "" {
		return errors.New("missing event source")
	}

	// If empty, the actor is presumed to be unknown - we do not record any events
	// for unknown actors.
	if event.Identifier == "" {
		oteltrace.SpanFromContext(spanCtx).
			RecordError(errors.New("event is missing actor identifier, discarding event"))
		return nil
	}

	// Always have metadata
	if event.Metadata == nil {
		event.Metadata = map[string]any{}
	}

	// HACK: Inject Sourcegraph actor that is held in the span context
	event.Metadata["sg.actor"] = sgactor.FromContext(spanCtx)

	// Inject trace metadata
	event.Metadata["trace_id"] = oteltrace.SpanContextFromContext(spanCtx).TraceID().String()

	metadata, err := json.Marshal(event.Metadata)
	if err != nil {
		return errors.Wrap(err, "marshaling metadata")
	}
	if err := l.tableInserter.Put(
		// Create a cancel-free context to avoid interrupting the log when
		// the parent context is cancelled.
		context.WithoutCancel(spanCtx),
		bigQueryEvent{
			Name:       string(event.Name),
			Source:     event.Source,
			Identifier: event.Identifier,
			Metadata:   json.RawMessage(metadata),
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
		Logger: &stdoutLogger{logger: logger.Scoped("events")},
	}
}

func (l *stdoutLogger) LogEvent(spanCtx context.Context, event Event) error {
	trace.Logger(spanCtx, l.logger).Debug("LogEvent",
		log.Object("event",
			log.String("name", string(event.Name)),
			log.String("source", event.Source),
			log.String("identifier", event.Identifier),
			log.String("metadata", fmt.Sprint(event.Metadata)),
		),
	)
	return nil
}
