package commitgraph

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewUpdater(uploadSvc UploadService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		uploadSvc.NewCommitGraphUpdater(
			ConfigInst.CommitGraphUpdateTaskInterval,
			ConfigInst.MaxAgeForNonStaleBranches,
			ConfigInst.MaxAgeForNonStaleTags,
		),
	}
}
