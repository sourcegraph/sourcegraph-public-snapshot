// Package compression handles compressing the number of data points that need to be searched for a code insight series.
//
// The purpose is to reduce the extremely large number of search queries that need to run to backfill a historical insight.
//
// An index of commits is used to understand which time frames actually contain changes in a given repository.
// The commit index comes with metadata for each repository that understands the time at which the index was most recently updated.
// It is relevant to understand whether the index can be considered up to date for a repository or not, otherwise
// frames could be filtered out that simply are not yet indexed and otherwise should be queried.
//
// The commit indexer also has the concept of a horizon, that is to say the farthest date at which indices are stored. This horizon
// does not necessarily correspond to the last commit in the repository (the repo could be much older) so the compression must also
// understand this.
//
// At a high level, the algorithm is as follows:
//
// * Given a series of time frames [1....N]:
// * Always include 1 (to establish a baseline at the max horizon so that last observations may be carried)
// * Never include N (let the indexed search handle this)
// * For each remaining frame, check if it has commit metadata that is up to date, and check if it has no commits. If so, throw out the frame
// * Otherwise, keep the frame
package compression

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/insights/priority"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/background/queryrunner"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/store"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type CommitFilter struct {
	store         CommitStore
	maxHistorical time.Time
}

type NoopFilter struct {
}

type Frame struct {
	From   time.Time
	To     time.Time
	Commit string
}

type DataFrameFilter interface {
	FilterFrames(ctx context.Context, frames []Frame, id api.RepoID) BackfillPlan
}

func NewHistoricalFilter(enabled bool, maxHistorical time.Time, db dbutil.DB) DataFrameFilter {
	if enabled {
		return &CommitFilter{
			store:         NewCommitStore(db),
			maxHistorical: maxHistorical,
		}
	}
	return &NoopFilter{}
}

func (n *NoopFilter) FilterFrames(ctx context.Context, frames []Frame, id api.RepoID) BackfillPlan {
	return uncompressedPlan(frames)
}

func uncompressedPlan(frames []Frame) BackfillPlan {
	executions := make([]*QueryExecution, 0, len(frames))
	for _, frame := range frames {
		executions = append(executions, &QueryExecution{RecordingTime: frame.From})
	}

	return BackfillPlan{
		Executions:  executions,
		RecordCount: len(executions),
	}
}

// FilterFrames will remove any data frames that can be safely skipped from a given frame set and for a given repository.
func (c *CommitFilter) FilterFrames(ctx context.Context, frames []Frame, id api.RepoID) BackfillPlan {
	include := make([]*QueryExecution, 0)
	// we will maintain a pointer to the most recent QueryExecution that we can update it's dependents
	var prev *QueryExecution
	var count int

	addToPlan := func(frame Frame, revhash string) {
		q := QueryExecution{RecordingTime: frame.From, Revision: revhash}
		include = append(include, &q)
		prev = &q
		count++
	}

	if len(frames) <= 1 {
		return uncompressedPlan(frames)
	}
	metadata, err := c.store.GetMetadata(ctx, id)
	if err != nil {
		// the commit index is considered optional so we can always fall back to every frame in this case
		return uncompressedPlan(frames)
	}

	// The first frame will always be included to establish a baseline measurement. This is important because
	// it is possible that the commit index will report zero commits because they may have happened beyond the
	// horizon of the indexer
	addToPlan(frames[0], "")
	for i := 1; i < len(frames); i++ {
		frame := frames[i]
		if metadata.LastIndexedAt.Before(frame.To) {
			// The commit indexer is not up to date enough to understand if this frame can be dropped
			addToPlan(frame, "")
			continue
		}

		commits, err := c.store.Get(ctx, id, frame.From, frame.To)
		if err != nil {
			log15.Error("insights: compression.go/FilterFrames unable to retrieve commits\n", "repo_id", id, "from", frame.From, "to", frame.To)
			addToPlan(frame, "")
			continue
		}
		// TODO(insights): record the commit here to save time having to look up which revhash we need since we already have it

		if len(commits) == 0 {
			// We have established that
			// 1. the commit index is sufficiently up to date
			// 2. this time range [from, to) doesn't have any commits
			// so we can skip this frame for this repo
			prev.SharedRecordings = append(prev.SharedRecordings, frame.From)
			count++
			continue
		} else {
			rev := commits[len(commits)-1]
			addToPlan(frame, string(rev.Commit))
		}
	}
	return BackfillPlan{Executions: include, RecordCount: count}
}

func (q *QueryExecution) Times() []time.Time {
	times := make([]time.Time, 0, q.RecordCount())

	times = append(times, q.RecordingTime)
	times = append(times, q.SharedRecordings...)
	return times
}

func (q *QueryExecution) RecordCount() int {
	return len(q.SharedRecordings) + 1
}

func (q *QueryExecution) ToRecording(seriesID string, repoName string, repoID api.RepoID, value float64) []store.RecordSeriesPointArgs {
	args := make([]store.RecordSeriesPointArgs, 0, q.RecordCount())
	base := store.RecordSeriesPointArgs{
		SeriesID: seriesID,
		Point: store.SeriesPoint{
			Time:  q.RecordingTime,
			Value: value,
		},
		RepoName: &repoName,
		RepoID:   &repoID,
	}
	args = append(args, base)
	for _, sharedTime := range q.SharedRecordings {
		arg := base
		arg.Point.Time = sharedTime
		args = append(args, arg)
	}

	return args
}

func (q *QueryExecution) ToQueueJob(seriesID string, query string, cost priority.Cost, jobPriority priority.Priority) *queryrunner.Job {
	return &queryrunner.Job{
		SeriesID:        seriesID,
		SearchQuery:     query,
		RecordTime:      &q.RecordingTime,
		Cost:            int(cost),
		Priority:        int(jobPriority),
		DependentFrames: q.SharedRecordings,
		State:           "queued",
	}
}

type BackfillPlan struct {
	Executions  []*QueryExecution
	RecordCount int
}

type QueryExecution struct {
	Revision         string
	RecordingTime    time.Time
	SharedRecordings []time.Time
}
