package backfill

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCommittedAtBackfiller(uploadSvc UploadService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		uploadSvc.NewCommittedAtBackfiller(
			ConfigInst.Interval,
			ConfigInst.BatchSize,
		),
	}
}
