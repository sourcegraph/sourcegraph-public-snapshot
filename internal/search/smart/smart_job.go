package smart

import (
	"context"
	"sort"
	"sync"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

func NewSmartJob(b query.Basic, newJob func(query.Basic) (job.Job, error)) (job.Job, error) {
	smartQuery, err := basicQueryToSmartQuery(b)
	if err != nil || smartQuery == nil {
		return nil, err
	}

	child, err := newJob(smartQuery.query)
	if err != nil {
		return nil, err
	}
	return &smartJob{child: child, patterns: smartQuery.patterns}, nil
}

type smartJob struct {
	child    job.Job
	patterns []string
}

func (j *smartJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	smartStream := newSmartStream(stream, j.patterns)
	return j.child.Run(ctx, clients, smartStream)
}

func (j *smartJob) Name() string {
	return "SmartJob"
}

func (j *smartJob) Fields(v job.Verbosity) (res []log.Field) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Printf("smart", ""),
		)
	}
	return res
}

func (j *smartJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *smartJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, fn)
	return &cp
}

// TODO: Can we somehow wait and aggregate _some amount_ of matches from multiple search events before sending them out?
// TODO: That way we could have at least a pseudo-ranking system.
func newSmartStream(parent streaming.Sender, patterns []string) streaming.Sender {
	var mux sync.Mutex
	return streaming.StreamFunc(func(e streaming.SearchEvent) {
		mux.Lock()

		validGroups := []matchGroup{}
		for _, r := range e.Results {
			fm, isFileMatch := r.(*result.FileMatch)
			if isFileMatch && len(fm.ChunkMatches) > 0 {
				fileScore := getFileScore(fm.Path, patterns)
				groups := groupChunkMatches(fm, fileScore, fm.ChunkMatches, float64(len(patterns)))

				for _, group := range groups {
					if group.IsValid() {
						validGroups = append(validGroups, group)
					}
				}
			}
		}

		sort.Slice(validGroups, func(i, j int) bool {
			return validGroups[i].Score() > validGroups[j].Score()
		})

		selected := e.Results[:0]
		// Flatten valid groups into a result stream (one match group per file).
		for _, group := range validGroups {
			selected = append(selected, &result.FileMatch{
				File:         group.fileMatch.File,
				ChunkMatches: group.group,
				LimitHit:     group.fileMatch.LimitHit,
			})
		}
		e.Results = selected

		mux.Unlock()
		parent.Send(e)
	})
}
