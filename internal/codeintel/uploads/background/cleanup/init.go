package cleanup

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewJanitor(backgroundJobs UploadServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewJanitor(
			ConfigInst.Interval,
			ConfigInst.UploadTimeout,
			ConfigInst.AuditLogMaxAge,
			ConfigInst.MinimumTimeSinceLastCheck,
			ConfigInst.CommitResolverBatchSize,
			ConfigInst.CommitResolverMaximumCommitLag,
		),
	}
}

func NewResetters(backgroundJobs UploadServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewUploadResetter(ConfigInst.Interval),
	}
}
