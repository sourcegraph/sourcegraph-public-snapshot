package events

import (
	"context"
	"encoding/json"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/sourcegraph/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codygateway"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var tracer = otel.GetTracerProvider().Tracer("cody-gateway/internal/events")

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
	return &bigQueryLogger{
		tableInserter: client.Dataset(dataset).Table(table).Inserter(),
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

	ctx, span := tracer.Start(backgroundContextWithSpan(spanCtx), "bigQueryLogger.LogEvent",
		trace.WithAttributes(attribute.String("source", event.Source), attribute.String("name", string(event.Name))))
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "failed to log event")
		}
		span.End()
	}()

	var metadata json.RawMessage
	if event.Metadata != nil {
		var err error
		metadata, err = json.Marshal(event.Metadata)
		if err != nil {
			return errors.Wrap(err, "marshaling metadata")
		}
	}

	if err := l.tableInserter.Put(
		ctx,
		bigQueryEvent{
			Name:       string(event.Name),
			Source:     event.Source,
			Identifier: event.Identifier,
			Metadata:   metadata,
			CreatedAt:  time.Now(),
		},
	); err != nil {
		return errors.Wrap(err, "inserting event")
	}
	return nil
}

type stdoutLogger struct {
	logger log.Logger
}

// NewStdoutLogger returns a new stdout event logger.
func NewStdoutLogger(logger log.Logger) Logger {
	return &stdoutLogger{logger: logger.Scoped("events", "event logger")}
}

func (l *stdoutLogger) LogEvent(spanCtx context.Context, event Event) error {
	// Not a terribly interesting trace, but useful to demo backgroundContextWithSpan
	_, span := tracer.Start(backgroundContextWithSpan(spanCtx), "stdoutLogger.LogEvent",
		trace.WithAttributes(
			attribute.String("source", event.Source),
			attribute.String("name", string(event.Name))))
	defer span.End()

	l.logger.Debug("LogEvent",
		log.Object("event",
			log.String("name", string(event.Name)),
			log.String("source", event.Source),
			log.String("identifier", event.Identifier),
		),
	)
	return nil
}

func backgroundContextWithSpan(ctx context.Context) context.Context {
	// NOTE: Using context.Background() because we still want to log the event in the
	// case of a request cancellation, we only want the parent span.
	ctx = trace.ContextWithSpan(context.Background(), trace.SpanFromContext(ctx))
	return ctx
}
