package jobutil

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/telemetry"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetrystore/teestore"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

// NewLogJob wraps a job with a LogJob, which records an event in the EventLogs table.
func NewLogJob(inputs *search.Inputs, child job.Job) job.Job {
	return &LogJob{
		child:  child,
		inputs: inputs,
	}
}

type LogJob struct {
	child  job.Job
	inputs *search.Inputs
}

func (l *LogJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := job.StartSpan(ctx, s, l)
	defer func() { finish(alert, err) }()

	start := time.Now()

	alert, err = l.child.Run(ctx, clients, s)

	duration := time.Since(start)

	l.logEvent(ctx, clients, duration)

	return alert, err
}

func (l *LogJob) Name() string {
	return "LogJob"
}

func (l *LogJob) Attributes(v job.Verbosity) (res []attribute.KeyValue) { return nil }

func (l *LogJob) Children() []job.Describer {
	return []job.Describer{l.child}
}

func (l *LogJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *l
	cp.child = job.Map(l.child, fn)
	return &cp
}

// logEvent records search durations in the event database. This function may
// only be called after a search result is performed, because it relies on the
// invariant that query and pattern error checking has already been performed.
func (l *LogJob) logEvent(ctx context.Context, clients job.RuntimeClients, duration time.Duration) {
	a := actor.FromContext(ctx)
	if !a.IsAuthenticated() || a.IsMockUser() {
		return
	}

	tr, ctx := trace.New(ctx, "LogSearchDuration")
	defer tr.End()

	var types []string
	resultTypes, _ := l.inputs.Query.StringValues(query.FieldType)
	for _, typ := range resultTypes {
		switch typ {
		case "repo", "symbol", "diff", "commit":
			types = append(types, typ)
		case "path":
			// Map type:path to file
			types = append(types, "file")
		case "file":
			types = append(types, l.inputs.PatternType.String())
		}
	}

	// Prefer the result type if it is a single type, otherwise use the pattern
	// type.
	action := ""
	if len(types) == 1 {
		action = types[0]
	} else {
		action = l.inputs.PatternType.String()
	}

	// New events that get exported: https://docs-legacy.sourcegraph.com/dev/background-information/telemetry
	events := telemetryrecorder.NewBestEffort(clients.Logger, clients.DB)
	// For now, do not tee into event_logs in telemetryrecorder - retain the
	// custom instrumentation of V1 events instead (usagestats.LogBackendEvent)
	ctx = teestore.WithoutV1(ctx)

	// Search.latencies
	events.Record(ctx, "search.latencies", telemetry.SafeAction(action), &telemetry.EventParameters{
		Metadata: telemetry.EventMetadata{
			"durationMs": float64(duration.Milliseconds()),
		},
	})
	// Legacy event
	value := fmt.Sprintf(`{"durationMs": %d}`, duration.Milliseconds())
	eventName := fmt.Sprintf("search.latencies.%s", action)
	//lint:ignore SA1019 existing usage of deprecated functionality. TODO: Use only the new V2 event instead.
	err := usagestats.LogBackendEvent(clients.DB, a.UID, deviceid.FromContext(ctx), eventName, json.RawMessage(value), json.RawMessage(value), featureflag.GetEvaluatedFlagSet(ctx), nil)
	if err != nil {
		clients.Logger.Warn("Could not log search latency", log.Error(err))
	}

	// Log usage of file:has.owners and select:file.owners
	q, err := query.ToBasicQuery(l.inputs.Query)
	if err != nil {
		return
	}

	if _, _, ok := isOwnershipSearch(q); ok {
		// New event
		events.Record(ctx, "search", "file.hasOwners", nil)
		//lint:ignore SA1019 existing usage of deprecated functionality. TODO: Use only the new V2 event instead.
		err := usagestats.LogBackendEvent(clients.DB, a.UID, deviceid.FromContext(ctx), "FileHasOwnerSearch", nil, nil, featureflag.GetEvaluatedFlagSet(ctx), nil)
		if err != nil {
			clients.Logger.Warn("Could not log use of file:has.owners", log.Error(err))
		}
	}

	if v, _ := q.ToParseTree().StringValue(query.FieldSelect); v != "" {
		if sp, err := filter.SelectPathFromString(v); err == nil && isSelectOwnersSearch(sp) {
			// New event
			events.Record(ctx, "search", "select.fileOwners", nil)
			//lint:ignore SA1019 existing usage of deprecated functionality. TODO: Use only the new V2 event instead.
			err := usagestats.LogBackendEvent(clients.DB, a.UID, deviceid.FromContext(ctx), "SelectFileOwnersSearch", nil, nil, featureflag.GetEvaluatedFlagSet(ctx), nil)
			if err != nil {
				clients.Logger.Warn("Could not log use of select:file.owners", log.Error(err))
			}
		}
	}
}
