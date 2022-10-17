package commitgraph

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewUpdater(backgroundJobs UploadServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewCommitGraphUpdater(
			ConfigInst.CommitGraphUpdateTaskInterval,
			ConfigInst.MaxAgeForNonStaleBranches,
			ConfigInst.MaxAgeForNonStaleTags,
		),
	}
}
