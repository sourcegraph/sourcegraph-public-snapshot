package backfill

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewCommittedAtBackfiller(backgroundJobs UploadServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewCommittedAtBackfiller(
			ConfigInst.Interval,
			ConfigInst.BatchSize,
		),
	}
}
