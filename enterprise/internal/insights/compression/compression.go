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

// uncompressedPlan returns a query plan that is completely uncompressed given an initial set of seed frames.
// This is primarily useful when there are scenarios in which compression cannot be used.
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
		log15.Error("unable to retrieve commit index metadata", "repo_id", id, "error", err)
		return uncompressedPlan(frames)
	}

	// The first frame will always be included to establish a baseline measurement. This is important because
	// it is possible that the commit index will report zero commits because they may have happened beyond the
	// horizon of the indexer
	addToPlan(frames[0], "")
	for i := 1; i < len(frames); i++ {
		previous := frames[i-1]
		frame := frames[i]
		if metadata.LastIndexedAt.Before(frame.To) {
			// The commit indexer is not up to date enough to understand if this frame can be dropped
			addToPlan(frame, "")
			continue
		}

		// We have to diff the commits in the previous frame to determine if we should query at the start of this frame
		commits, err := c.store.Get(ctx, id, previous.From, previous.To)
		if err != nil {
			log15.Error("insights: compression.go/FilterFrames unable to retrieve commits\n", "repo_id", id, "from", frame.From, "to", frame.To)
			addToPlan(frame, "")
			continue
		}
		if len(commits) == 0 {
			// We have established that
			// 1. the commit index is sufficiently up to date
			// 2. this time range [from, to) doesn't have any commits
			// so we can skip this frame for this repo
			log15.Info("insights: skipping query based on no commits", "for_time", frame.From, "repo_id", id)
			prev.SharedRecordings = append(prev.SharedRecordings, frame.From)
			count++
			continue
		} else {
			rev := commits[0]
			log15.Info("insights: generating query with commit index revision", "rev", rev, "for_time", frame.From, "repo_id", id)
			// as a small optimization we are collecting this revhash here since we already know this is
			// the revision for which we need to query against
			addToPlan(frame, string(rev.Commit))
		}
	}
	return BackfillPlan{Executions: include, RecordCount: count}
}

// RecordCount returns the total count of data points that will be generated by this execution.
func (q *QueryExecution) RecordCount() int {
	return len(q.SharedRecordings) + 1
}

// ToRecording converts the query execution into a slice of recordable data points, each sharing the same value.
func (q *QueryExecution) ToRecording(seriesID string, repoName string, repoID api.RepoID, value float64) []store.RecordSeriesPointArgs {
	args := make([]store.RecordSeriesPointArgs, 0, q.RecordCount())
	base := store.RecordSeriesPointArgs{
		SeriesID: seriesID,
		Point: store.SeriesPoint{
			Time:  q.RecordingTime,
			Value: value,
		},
		RepoName:    &repoName,
		RepoID:      &repoID,
		PersistMode: store.RecordMode,
	}
	args = append(args, base)
	for _, sharedTime := range q.SharedRecordings {
		arg := base
		arg.Point.Time = sharedTime
		args = append(args, arg)
	}

	return args
}

// ToQueueJob converts the query execution into a queueable job with it's relevant dependent times.
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

// BackfillPlan is a rudimentary query plan. It provides a simple mechanism to store executable nodes
// to backfill an insight series.
type BackfillPlan struct {
	Executions  []*QueryExecution
	RecordCount int
}

// QueryExecution represents a node of an execution plan that should be queried against Sourcegraph.
// It can have dependent time points that will inherit the same value as the exemplar point
// once the query is executed and resolved.
type QueryExecution struct {
	Revision         string
	RecordingTime    time.Time
	SharedRecordings []time.Time
}
