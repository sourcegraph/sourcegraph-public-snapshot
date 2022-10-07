package cleanup

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewJanitor(uploadSvc UploadService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		uploadSvc.NewJanitor(
			ConfigInst.Interval,
			ConfigInst.UploadTimeout,
			ConfigInst.AuditLogMaxAge,
			ConfigInst.MinimumTimeSinceLastCheck,
			ConfigInst.CommitResolverBatchSize,
			ConfigInst.CommitResolverMaximumCommitLag,
		),
	}
}

func NewResetters(uploadSvc UploadService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		uploadSvc.NewUploadResetter(ConfigInst.Interval),
	}
}
