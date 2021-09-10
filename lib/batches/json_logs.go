package batches

import "time"

type LogEvent struct {
	Operation LogEventOperation `json:"operation"`

	Timestamp time.Time `json:"timestamp"`

	Status   LogEventStatus         `json:"status"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

type LogEventOperation string

const (
	LogEventOperationParsingBatchSpec          LogEventOperation = "PARSING_BATCH_SPEC"
	LogEventOperationResolvingNamespace        LogEventOperation = "RESOLVING_NAMESPACE"
	LogEventOperationPreparingDockerImages     LogEventOperation = "PREPARING_DOCKER_IMAGES"
	LogEventOperationDeterminingWorkspaceType  LogEventOperation = "DETERMINING_WORKSPACE_TYPE"
	LogEventOperationResolvingRepositories     LogEventOperation = "RESOLVING_REPOSITORIES"
	LogEventOperationDeterminingWorkspaces     LogEventOperation = "DETERMINING_WORKSPACES"
	LogEventOperationCheckingCache             LogEventOperation = "CHECKING_CACHE"
	LogEventOperationExecutingTasks            LogEventOperation = "EXECUTING_TASKS"
	LogEventOperationLogFileKept               LogEventOperation = "LOG_FILE_KEPT"
	LogEventOperationUploadingChangesetSpecs   LogEventOperation = "UPLOADING_CHANGESET_SPECS"
	LogEventOperationCreatingBatchSpec         LogEventOperation = "CREATING_BATCH_SPEC"
	LogEventOperationApplyingBatchSpec         LogEventOperation = "APPLYING_BATCH_SPEC"
	LogEventOperationBatchSpecExecution        LogEventOperation = "BATCH_SPEC_EXECUTION"
	LogEventOperationExecutingTask             LogEventOperation = "EXECUTING_TASK"
	LogEventOperationTaskBuildChangesetSpecs   LogEventOperation = "TASK_BUILD_CHANGESET_SPECS"
	LogEventOperationTaskDownloadingArchive    LogEventOperation = "TASK_DOWNLOADING_ARCHIVE"
	LogEventOperationTaskInitializingWorkspace LogEventOperation = "TASK_INITIALIZING_WORKSPACE"
	LogEventOperationTaskSkippingSteps         LogEventOperation = "TASK_SKIPPING_STEPS"
	LogEventOperationTaskStepSkipped           LogEventOperation = "TASK_STEP_SKIPPED"
	LogEventOperationTaskPreparingStep         LogEventOperation = "TASK_PREPARING_STEP"
	LogEventOperationTaskStep                  LogEventOperation = "TASK_STEP"
	LogEventOperationTaskCalculatingDiff       LogEventOperation = "TASK_CALCULATING_DIFF"
)

type LogEventStatus string

const (
	LogEventStatusStarted  LogEventStatus = "STARTED"
	LogEventStatusSuccess  LogEventStatus = "SUCCESS"
	LogEventStatusFailure  LogEventStatus = "FAILURE"
	LogEventStatusProgress LogEventStatus = "PROGRESS"
)
