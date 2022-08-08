package keyword

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

func NewKeywordSearchJob(b query.Basic, newJob func(query.Basic) (job.Job, error)) (job.Job, error) {
	keywordQuery, err := basicQueryToKeywordQuery(b)
	if err != nil || keywordQuery == nil {
		return nil, err
	}

	child, err := newJob(keywordQuery.query)
	if err != nil {
		return nil, err
	}
	return &keywordSearchJob{child: child, patterns: keywordQuery.patterns}, nil
}

type keywordSearchJob struct {
	child    job.Job
	patterns []string
}

func (j *keywordSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	_, ctx, stream, finish := job.StartSpan(ctx, stream, j)
	defer func() { finish(alert, err) }()

	// TODO(novoselrok): Use NewBatchingStream to batch the events before processing them.
	keywordSearchStream := newKeywordSearchStream(stream, j.patterns)
	return j.child.Run(ctx, clients, keywordSearchStream)
}

func (j *keywordSearchJob) Name() string {
	return "KeywordSearchJob"
}

func (j *keywordSearchJob) Fields(v job.Verbosity) (res []log.Field) {
	switch v {
	case job.VerbosityMax:
		fallthrough
	case job.VerbosityBasic:
		res = append(res,
			trace.Printf("keyword", ""),
		)
	}
	return res
}

func (j *keywordSearchJob) Children() []job.Describer {
	return []job.Describer{j.child}
}

func (j *keywordSearchJob) MapChildren(fn job.MapFunc) job.Job {
	cp := *j
	cp.child = job.Map(j.child, fn)
	return &cp
}

func newKeywordSearchStream(parent streaming.Sender, patterns []string) streaming.Sender {
	var mux sync.Mutex
	return streaming.StreamFunc(func(e streaming.SearchEvent) {
		mux.Lock()

		relevantGroups := []matchGroup{}
		for _, r := range e.Results {
			fm, isFileMatch := r.(*result.FileMatch)
			if isFileMatch && len(fm.ChunkMatches) > 0 {
				fileScore := getFileScore(fm.Path, patterns)
				groups := groupChunkMatches(fm, fileScore, fm.ChunkMatches, float64(len(patterns)))

				for _, group := range groups {
					if group.IsRelevant() {
						relevantGroups = append(relevantGroups, group)
					}
				}
			}
		}

		sort.Slice(relevantGroups, func(i, j int) bool {
			return relevantGroups[i].Score() > relevantGroups[j].Score()
		})

		selected := e.Results[:0]
		// Flatten valid groups into a result stream (one match group per file).
		for _, group := range relevantGroups {
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
