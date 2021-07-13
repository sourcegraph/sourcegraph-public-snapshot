package main

import (
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
)

type batchExecUI interface {
	ParsingBatchSpec()
	ParsingBatchSpecSuccess()
	ParsingBatchSpecFailure(error)

	ResolvingNamespace()
	ResolvingNamespaceSuccess(namespace string)

	PreparingContainerImages()
	PreparingContainerImagesProgress(percent float64)
	PreparingContainerImagesSuccess()

	DeterminingWorkspaceCreatorType()
	DeterminingWorkspaceCreatorTypeSuccess(wt workspace.CreatorType)

	ResolvingRepositories()
	ResolvingRepositoriesDone(repos []*graphql.Repository, unsupported batches.UnsupportedRepoSet, ignored batches.IgnoredRepoSet)

	DeterminingWorkspaces()
	DeterminingWorkspacesSuccess(num int)

	CheckingCache()
	CheckingCacheSuccess(cachedSpecsFound int, tasksToExecute int)

	ExecutingTasks(verbose bool, parallelism int) func(ts []*executor.TaskStatus)
	ExecutingTasksSkippingErrors(err error)
	ExecutingTasksSuccess()

	LogFilesKept(files []string)

	NoChangesetSpecs()
	UploadingChangesetSpecs(num int)
	UploadingChangesetSpecsProgress(done, total int)
	UploadingChangesetSpecsSuccess()

	CreatingBatchSpec()
	CreatingBatchSpecSuccess()
	CreatingBatchSpecError(err error) error

	PreviewBatchSpec(previewURL string)

	ApplyingBatchSpec()
	ApplyingBatchSpecSuccess(batchChangeURL string)

	ExecutionError(error)
}
