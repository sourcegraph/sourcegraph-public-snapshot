package compression

import (
	"context"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
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
	FilterFrames(ctx context.Context, frames []Frame, id api.RepoID) []Frame
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

func (n *NoopFilter) FilterFrames(ctx context.Context, frames []Frame, id api.RepoID) []Frame {
	return frames
}

// FilterFrames will remove any data frames that can be safely skipped from a given frame set and for a given repository.
func (c *CommitFilter) FilterFrames(ctx context.Context, frames []Frame, id api.RepoID) []Frame {
	if len(frames) <= 1 {
		return frames
	}

	metadata, err := c.store.GetMetadata(ctx, id)
	if err != nil {
		// the commit index is considered optional so we can always fall back to every frame in this case
		return frames
	}

	include := make([]Frame, 0)
	// The first frame will always be included to establish a baseline measurement. This is important because
	// it is possible that the commit index will report zero commits because they may have happened beyond the
	// horizon of the indexer
	include = append(include, frames[0])

	// The last frame will always be excluded because this is the frame closest to now. In this case, we will allow
	// the most recent sample to be resolved from the 'present day recorder' insight_enqueuer.go
	for i := 1; i < len(frames)-1; i++ {
		frame := frames[i]

		if metadata.LastIndexedAt.Before(frame.To) {
			// The commit indexer is not up to date enough to understand if this frame can be dropped
			include = append(include, frame)
			continue
		}

		commits, err := c.store.Get(ctx, id, frame.From, frame.To)
		if err != nil {
			log15.Error("insights: compression.go/FilterFrames unable to retrieve commits\n", "repo_id", id, "from", frame.From, "to", frame.To)
			include = append(include, frame)
			continue
		}
		// TODO(insights): record the commit here to save time having to look up which revhash we need since we already have it

		if len(commits) == 0 {
			// We have established that
			// 1. the commit index is sufficiently up to date
			// 2. this time range [from, to) doesn't have any commits
			// so we can skip this frame for this repo
			continue
		} else {
			include = append(include, frame)
		}
	}
	return include
}
