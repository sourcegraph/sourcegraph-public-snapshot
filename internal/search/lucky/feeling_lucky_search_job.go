package lucky

import (
	"context"
	"fmt"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	alertobserver "github.com/sourcegraph/sourcegraph/internal/search/alert"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/limits"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// autoQuery is an automatically generated query with associated data (e.g., description).
type autoQuery struct {
	description string
	query       query.Basic
}

// newJob is a function that converts a query to a job, and one which lucky
// search expects in order to function. This function corresponds to
// `jobutil.NewBasicJob` normally (we can't call it directly for circular
// dependencies), and otherwise abstracts job creation for tests.
type newJob func(query.Basic) (job.Job, error)

// NewFeelingLuckySearchJob creates generators for opportunistic search queries
// that apply various rules, transforming the original input plan into various
// queries that alter its interpretation (e.g., search literally for quotes or
// not, attempt to search the pattern as a regexp, and so on). There is no
// random choice when applying rules.
func NewFeelingLuckySearchJob(initialJob job.Job, newJob newJob, plan query.Plan) *FeelingLuckySearchJob {
	generators := make([]next, 0, len(plan))
	for _, b := range plan {
		generators = append(generators, NewGenerator(b, rulesNarrow, rulesWiden))
	}

	newGeneratedJob := func(autoQ *autoQuery) job.Job {
		child, err := newJob(autoQ.query)
		if err != nil {
			return nil
		}

		notifier := &notifier{autoQuery: autoQ}

		return &generatedSearchJob{
			Child:           child,
			NewNotification: notifier.New,
		}
	}

	return &FeelingLuckySearchJob{
		initialJob:      initialJob,
		generators:      generators,
		newGeneratedJob: newGeneratedJob,
	}
}

// FeelingLuckySearchJob represents a lucky search. Note `newGeneratedJob`
// returns a job given an autoQuery. It is a function so that generated queries
// can be composed at runtime (with auto queries that dictate runtime control
// flow) with static inputs (search inputs), while not exposing static inputs.
type FeelingLuckySearchJob struct {
	initialJob      job.Job
	generators      []next
	newGeneratedJob func(*autoQuery) job.Job
}

// Do not run autogenerated queries if RESULT_THRESHOLD results exist on the original query.
const RESULT_THRESHOLD = limits.DefaultMaxSearchResultsStreaming

func (f *FeelingLuckySearchJob) Run(ctx context.Context, clients job.RuntimeClients, parentStream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, parentStream, finish := job.StartSpan(ctx, parentStream, f)
	defer func() { finish(alert, err) }()

	dedupingStream := streaming.NewDedupingStream(parentStream)
	// Count stream results to know whether to run generated queries
	stream := streaming.NewResultCountingStream(dedupingStream)

	var maxAlerter search.MaxAlerter
	var errs errors.MultiError
	alert, err = f.initialJob.Run(ctx, clients, stream)
	if err != nil {
		return alert, err
	}
	maxAlerter.Add(alert)

	initialResultSetSize := stream.Count()
	if initialResultSetSize >= RESULT_THRESHOLD {
		return alert, err
	}

	var luckyAlertType alertobserver.LuckyAlertType
	if initialResultSetSize == 0 {
		luckyAlertType = alertobserver.LuckyAlertPure
	} else {
		luckyAlertType = alertobserver.LuckyAlertAdded
	}
	generated := &alertobserver.ErrLuckyQueries{Type: luckyAlertType, ProposedQueries: []*search.QueryDescription{}}
	var autoQ *autoQuery
	for _, next := range f.generators {
		for next != nil {
			autoQ, next = next()
			j := f.newGeneratedJob(autoQ)
			if j == nil {
				// Generated an invalid job with this query, just continue.
				continue
			}
			alert, err = j.Run(ctx, clients, stream)
			if stream.Count()-initialResultSetSize >= RESULT_THRESHOLD {
				// We've sent additional results up to the maximum bound. Let's stop here.
				var lErr *alertobserver.ErrLuckyQueries
				if errors.As(err, &lErr) {
					generated.ProposedQueries = append(generated.ProposedQueries, lErr.ProposedQueries...)
				}
				if len(generated.ProposedQueries) > 0 {
					errs = errors.Append(errs, generated)
				}
				return maxAlerter.Alert, errs
			}

			var lErr *alertobserver.ErrLuckyQueries
			if errors.As(err, &lErr) {
				// collected generated queries, we'll add it after this loop is done running.
				generated.ProposedQueries = append(generated.ProposedQueries, lErr.ProposedQueries...)
			} else {
				errs = errors.Append(errs, err)
			}

			maxAlerter.Add(alert)
		}
	}

	if len(generated.ProposedQueries) > 0 {
		errs = errors.Append(errs, generated)
	}
	return maxAlerter.Alert, errs
}

func (f *FeelingLuckySearchJob) Name() string {
	return "FeelingLuckySearchJob"
}

func (f *FeelingLuckySearchJob) Fields(job.Verbosity) []log.Field { return nil }

func (f *FeelingLuckySearchJob) Children() []job.Describer {
	return []job.Describer{f.initialJob}
}

func (f *FeelingLuckySearchJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *f
	cp.initialJob = job.Map(f.initialJob, fn)
	return &cp
}

// generatedSearchJob represents a generated search at run time. Note
// `NewNotification` returns the query notifications (encoded as error) given
// the result count of the job. It is a function so that notifications can be
// composed at runtime (with result counts) with static inputs (query string),
// while not exposing static inputs.
type generatedSearchJob struct {
	Child           job.Job
	NewNotification func(count int) error
}

func (g *generatedSearchJob) Run(ctx context.Context, clients job.RuntimeClients, parentStream streaming.Sender) (*search.Alert, error) {
	stream := streaming.NewResultCountingStream(parentStream)
	alert, err := g.Child.Run(ctx, clients, stream)
	resultCount := stream.Count()
	if resultCount == 0 {
		return nil, nil
	}

	if ctx.Err() != nil {
		notification := g.NewNotification(resultCount)
		return alert, errors.Append(err, notification)
	}

	notification := g.NewNotification(resultCount)
	if err != nil {
		return alert, errors.Append(err, notification)
	}

	return alert, notification
}

func (g *generatedSearchJob) Name() string {
	return "GeneratedSearchJob"
}

func (g *generatedSearchJob) Children() []job.Describer { return []job.Describer{g.Child} }

func (g *generatedSearchJob) Fields(job.Verbosity) []log.Field { return nil }

func (g *generatedSearchJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *g
	cp.Child = job.Map(g.Child, fn)
	return &cp
}

// notifier stores static values that should not be exposed to runtime concerns.
// notifier exposes a method `New` for constructing notifications that require
// runtime information.
type notifier struct {
	*autoQuery
}

func (n *notifier) New(count int) error {
	var resultCountString string
	if count == limits.DefaultMaxSearchResultsStreaming {
		resultCountString = fmt.Sprintf("%d+ results", count)
	} else if count == 1 {
		resultCountString = fmt.Sprintf("1 result")
	} else {
		resultCountString = fmt.Sprintf("%d additional results", count)
	}
	annotations := make(map[search.AnnotationName]string)
	annotations[search.ResultCount] = resultCountString

	return &alertobserver.ErrLuckyQueries{
		ProposedQueries: []*search.QueryDescription{{
			Description: n.description,
			Annotations: map[search.AnnotationName]string{
				search.ResultCount: resultCountString,
			},
			Query:       query.StringHuman(n.query.ToParseTree()),
			PatternType: query.SearchTypeLucky,
		}},
	}
}
