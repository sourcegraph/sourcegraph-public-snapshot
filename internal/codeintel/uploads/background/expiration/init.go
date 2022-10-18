package expiration

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewExpirationTasks(backgroundJobs UploadServiceBackgroundJobs) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		backgroundJobs.NewUploadExpirer(
			ConfigInst.ExpirerInterval,
			ConfigInst.RepositoryProcessDelay,
			ConfigInst.RepositoryBatchSize,
			ConfigInst.UploadProcessDelay,
			ConfigInst.UploadBatchSize,
			ConfigInst.CommitBatchSize,
			ConfigInst.PolicyBatchSize,
		),
		backgroundJobs.NewReferenceCountUpdater(
			ConfigInst.ReferenceCountUpdaterInterval,
			ConfigInst.ReferenceCountUpdaterBatchSize,
		),
	}
}
