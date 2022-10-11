package expiration

import (
	"github.com/sourcegraph/sourcegraph/internal/goroutine"
)

func NewExpirationTasks(uploadSvc UploadService) []goroutine.BackgroundRoutine {
	return []goroutine.BackgroundRoutine{
		uploadSvc.NewUploadExpirer(
			ConfigInst.ExpirerInterval,
			ConfigInst.RepositoryProcessDelay,
			ConfigInst.RepositoryBatchSize,
			ConfigInst.UploadProcessDelay,
			ConfigInst.UploadBatchSize,
			ConfigInst.CommitBatchSize,
			ConfigInst.PolicyBatchSize,
		),
		uploadSvc.NewReferenceCountUpdater(
			ConfigInst.ReferenceCountUpdaterInterval,
			ConfigInst.ReferenceCountUpdaterBatchSize,
		),
	}
}
