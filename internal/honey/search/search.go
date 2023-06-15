package search

import (
	"context"

	oteltrace "go.opentelemetry.io/otel/trace"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/honey"
)

type SearchEventArgs struct {
	OriginalQuery string
	Typ           string
	Source        string
	Status        string
	AlertType     string
	DurationMs    int64
	LatencyMs     *int64
	ResultSize    int
	Error         error
}

// SearchEvent returns a honey event for the dataset "search".
func SearchEvent(ctx context.Context, args SearchEventArgs) honey.Event {
	act := actor.FromContext(ctx)
	ev := honey.NewEvent("search")
	ev.AddField("query", args.OriginalQuery)
	ev.AddField("actor_uid", act.UID)
	ev.AddField("actor_internal", act.Internal)
	ev.AddField("type", args.Typ)
	ev.AddField("source", args.Source)
	ev.AddField("status", args.Status)
	ev.AddField("alert_type", args.AlertType)
	ev.AddField("duration_ms", args.DurationMs)
	ev.AddField("latency_ms", args.LatencyMs)
	ev.AddField("result_size", args.ResultSize)
	if args.Error != nil {
		ev.AddField("error", args.Error.Error())
	}
	if span := oteltrace.SpanFromContext(ctx); span != nil {
		spanContext := span.SpanContext()
		ev.AddField("trace_id", spanContext.TraceID())
		ev.AddField("span_id", spanContext.SpanID())
	}

	return ev
}
