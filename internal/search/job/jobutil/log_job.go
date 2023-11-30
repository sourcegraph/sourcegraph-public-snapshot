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
	"github.com/sourcegraph/sourcegraph/internal/telemetry/teestore"
	"github.com/sourcegraph/sourcegraph/internal/telemetry/telemetryrecorder"
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
			switch {
			case l.inputs.PatternType == query.SearchTypeStandard:
				types = append(types, "standard")
			case l.inputs.PatternType == query.SearchTypeStructural:
				types = append(types, "structural")
			case l.inputs.PatternType == query.SearchTypeLiteral:
				types = append(types, "literal")
			case l.inputs.PatternType == query.SearchTypeRegex:
				types = append(types, "regexp")
			case l.inputs.PatternType == query.SearchTypeLucky:
				types = append(types, "lucky")
			}
		}
	}

	// Don't record composite searches that specify more than one type:
	// because we can't break down the search timings into multiple
	// categories.
	if len(types) > 1 {
		return
	}

	q, err := query.ToBasicQuery(l.inputs.Query)
	if err != nil {
		// Can't convert to a basic query, can't guarantee accurate reporting.
		return
	}
	if !query.IsPatternAtom(q) {
		// Not an atomic pattern, can't guarantee accurate reporting.
		return
	}

	// If no type: was explicitly specified, infer the result type.
	if len(types) == 0 {
		// If a pattern was specified, a content search happened.
		if q.IsLiteral() {
			types = append(types, "literal")
		} else if q.IsRegexp() {
			types = append(types, "regexp")
		} else if q.IsStructural() {
			types = append(types, "structural")
		} else if l.inputs.Query.Exists(query.FieldFile) {
			// No search pattern specified and file: is specified.
			types = append(types, "file")
		} else {
			// No search pattern or file: is specified, assume repo.
			// This includes accounting for searches of fields that
			// specify repohasfile: and repohascommitafter:.
			types = append(types, "repo")
		}
	}
	// Only log the time if we successfully resolved one search type.
	if len(types) == 1 {
		// New events that get exported: https://docs.sourcegraph.com/dev/background-information/telemetry
		events := telemetryrecorder.NewBestEffort(clients.Logger, clients.DB)
		// For now, do not tee into event_logs in telemetryrecorder - retain the
		// custom instrumentation of V1 events instead (usagestats.LogBackendEvent)
		ctx = teestore.WithoutV1(ctx)

		a := actor.FromContext(ctx)
		if a.IsAuthenticated() && !a.IsMockUser() { // Do not log in tests
			// New event
			events.Record(ctx, "search.latencies", telemetry.Action(types[0]), &telemetry.EventParameters{
				Metadata: telemetry.EventMetadata{
					"durationMs": telemetry.Number(duration.Milliseconds()),
					// TODO: Maybe make FromSessionCookie a first-class data
					// point in (telemetrygateway.proto).EventUser if more use
					// cases surface.
					"actor.fromSessionCookie": telemetry.Bool(a.FromSessionCookie),
				},
			})
			// Legacy event
			value := fmt.Sprintf(`{"durationMs": %d}`, duration.Milliseconds())
			eventName := fmt.Sprintf("search.latencies.%s", types[0])
			err := usagestats.LogBackendEvent(clients.DB, a.UID, deviceid.FromContext(ctx), eventName, json.RawMessage(value), json.RawMessage(value), featureflag.GetEvaluatedFlagSet(ctx), nil)
			if err != nil {
				clients.Logger.Warn("Could not log search latency", log.Error(err))
			}

			if _, _, ok := isOwnershipSearch(q); ok {
				// New event
				events.Record(ctx, "search", "file.hasOwners", nil)
				// Legacy event
				err := usagestats.LogBackendEvent(clients.DB, a.UID, deviceid.FromContext(ctx), "FileHasOwnerSearch", nil, nil, featureflag.GetEvaluatedFlagSet(ctx), nil)
				if err != nil {
					clients.Logger.Warn("Could not log use of file:has.owners", log.Error(err))
				}
			}

			if v, _ := q.ToParseTree().StringValue(query.FieldSelect); v != "" {
				if sp, err := filter.SelectPathFromString(v); err == nil && isSelectOwnersSearch(sp) {
					// New event
					events.Record(ctx, "search", "select.fileOwners", nil)
					// Legacy event
					err := usagestats.LogBackendEvent(clients.DB, a.UID, deviceid.FromContext(ctx), "SelectFileOwnersSearch", nil, nil, featureflag.GetEvaluatedFlagSet(ctx), nil)
					if err != nil {
						clients.Logger.Warn("Could not log use of select:file.owners", log.Error(err))
					}
				}
			}
		}
	}
}
