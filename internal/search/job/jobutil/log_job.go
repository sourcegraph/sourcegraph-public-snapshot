package jobutil

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	otlog "github.com/opentracing/opentracing-go/log"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/actor"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/deviceid"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/internal/honey"
	searchhoney "github.com/sourcegraph/sourcegraph/internal/honey/search"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

func NewLogJob(child job.Job) job.Job {
	return &LogJob{
		child: child,
	}
}

type LogJob struct {
	child  job.Job
	inputs search.Inputs
}

func (l *LogJob) Run(ctx context.Context, clients job.RuntimeClients, s streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, s, finish := job.StartSpan(ctx, s, l)
	defer func() { finish(alert, err) }()

	start := time.Now()

	statsStream := streaming.NewStatsObservingStream(s)
	resultStream := streaming.NewResultCountingStream(statsStream)
	alert, err = l.child.Run(ctx, clients, resultStream)

	duration := time.Since(start)

	l.logBatch(ctx, clients, l.inputs, statsStream.Stats, resultStream.Count(), duration, alert, err)

	return alert, err
}

func (l *LogJob) Name() string {
	return "LogJob"
}

func (l *LogJob) Fields(v job.Verbosity) (res []otlog.Field) { return nil }

func (l *LogJob) Children() []job.Describer {
	return []job.Describer{l.child}
}

func (l *LogJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *l
	cp.child = job.Map(l.child, fn)
	return &cp
}

func (l *LogJob) logBatch(
	ctx context.Context,
	clients job.RuntimeClients,
	inputs search.Inputs,
	stats streaming.Stats,
	resultCount int,
	duration time.Duration,
	alert *search.Alert,
	err error,
) {
	var wg sync.WaitGroup
	defer wg.Wait()
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.logSearchDuration(ctx, clients, inputs, duration)
	}()

	var status, alertType string
	status = determineStatusForLogs(alert, stats, err)
	if alert != nil {
		alertType = alert.PrometheusType
	}
	requestSource := string(trace.RequestSource(ctx))
	requestName := trace.GraphQLRequestName(ctx)
	logPrometheusBatch(status, alertType, requestSource, requestName, duration)

	isSlow := duration > slowSearchesThreshold()
	if honey.Enabled() || isSlow {
		ev := searchhoney.SearchEvent(ctx, searchhoney.SearchEventArgs{
			OriginalQuery: inputs.OriginalQuery,
			Typ:           requestName,
			Source:        requestSource,
			Status:        status,
			AlertType:     alertType,
			DurationMs:    duration.Milliseconds(),
			ResultSize:    resultCount,
			Error:         err,
		})

		_ = ev.Send()

		if isSlow {
			clients.Logger.Warn("slow search request",
				log.String("query", inputs.OriginalQuery),
				log.String("type", requestName),
				log.String("source", requestSource),
				log.String("status", status),
				log.String("alertType", alertType),
				log.Int64("durationMs", duration.Milliseconds()),
				log.Int("resultSize", resultCount),
				log.String("error", err.Error()),
			)
		}
	}
}

// logSearchDuration records search durations in the event database. This
// function may only be called after a search result is performed, because it
// relies on the invariant that query and pattern error checking has already
// been performed.
func (l *LogJob) logSearchDuration(ctx context.Context, clients job.RuntimeClients, si search.Inputs, duration time.Duration) {
	tr, ctx := trace.New(ctx, "LogSearchDuration", "")
	defer tr.Finish()
	var types []string
	resultTypes, _ := si.Query.StringValues(query.FieldType)
	for _, typ := range resultTypes {
		switch typ {
		case "repo", "symbol", "diff", "commit":
			types = append(types, typ)
		case "path":
			// Map type:path to file
			types = append(types, "file")
		case "file":
			switch {
			case si.PatternType == query.SearchTypeStandard:
				types = append(types, "standard")
			case si.PatternType == query.SearchTypeStructural:
				types = append(types, "structural")
			case si.PatternType == query.SearchTypeLiteral:
				types = append(types, "literal")
			case si.PatternType == query.SearchTypeRegex:
				types = append(types, "regexp")
			case si.PatternType == query.SearchTypeLucky:
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

	q, err := query.ToBasicQuery(si.Query)
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
		} else if si.Query.Exists(query.FieldFile) {
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
		a := actor.FromContext(ctx)
		if a.IsAuthenticated() && !a.IsMockUser() { // Do not log in tests
			value := fmt.Sprintf(`{"durationMs": %d}`, duration.Milliseconds())
			eventName := fmt.Sprintf("search.latencies.%s", types[0])
			err := usagestats.LogBackendEvent(clients.DB, a.UID, deviceid.FromContext(ctx), eventName, json.RawMessage(value), json.RawMessage(value), featureflag.GetEvaluatedFlagSet(ctx), nil)
			if err != nil {
				clients.Logger.Warn("Could not log search latency", log.Error(err))
			}
		}
	}
}

var (
	searchResponseCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "src_graphql_search_response",
		Help: "Number of searches that have ended in the given status (success, error, timeout, partial_timeout).",
	}, []string{"status", "alert_type", "source", "request_name"})

	searchLatencyHistogram = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "src_search_response_latency_seconds",
		Help:    "Search response latencies in seconds that have ended in the given status (success, error, timeout, partial_timeout).",
		Buckets: []float64{0.01, 0.02, 0.05, 0.1, 0.2, 0.5, 1, 2, 5, 10, 15, 20, 30},
	}, []string{"status", "alert_type", "source", "request_name"})
)

func logPrometheusBatch(status, alertType, requestSource, requestName string, elapsed time.Duration) {
	searchResponseCounter.WithLabelValues(
		status,
		alertType,
		requestSource,
		requestName,
	).Inc()

	searchLatencyHistogram.WithLabelValues(
		status,
		alertType,
		requestSource,
		requestName,
	).Observe(elapsed.Seconds())
}

// determineStatusForLogs determines the final status of a search for logging
// purposes.
func determineStatusForLogs(alert *search.Alert, stats streaming.Stats, err error) string {
	switch {
	case err == context.DeadlineExceeded:
		return "timeout"
	case err != nil:
		return "error"
	case stats.Status.All(search.RepoStatusTimedout) && stats.Status.Len() == len(stats.Repos):
		return "timeout"
	case stats.Status.Any(search.RepoStatusTimedout):
		return "partial_timeout"
	case alert != nil:
		return "alert"
	default:
		return "success"
	}
}

// slowSearchesThreshold returns the minimum duration configured in site
// settings for logging slow searches.
func slowSearchesThreshold() time.Duration {
	ms := conf.Get().ObservabilityLogSlowSearches
	if ms == 0 {
		return time.Duration(math.MaxInt64)
	}
	return time.Duration(ms) * time.Millisecond
}
