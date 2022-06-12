package batches

import (
	"encoding/json"
	"time"

	"github.com/sourcegraph/sourcegraph/lib/batches/execution"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type LogEvent struct {
	Operation LogEventOperation `json:"operation"`

	Timestamp time.Time `json:"timestamp"`

	Status   LogEventStatus `json:"status"`
	Metadata any            `json:"metadata,omitempty"`
}

type logEventJSON struct {
	Operation LogEventOperation `json:"operation"`
	Timestamp time.Time         `json:"timestamp"`
	Status    LogEventStatus    `json:"status"`
}

func (l *LogEvent) UnmarshalJSON(data []byte) error {
	var j *logEventJSON
	if err := json.Unmarshal(data, &j); err != nil {
		return err
	}
	l.Operation = j.Operation
	l.Timestamp = j.Timestamp
	l.Status = j.Status

	switch l.Operation {
	case LogEventOperationParsingBatchSpec:
		l.Metadata = new(ParsingBatchSpecMetadata)
	case LogEventOperationResolvingNamespace:
		l.Metadata = new(ResolvingNamespaceMetadata)
	case LogEventOperationPreparingDockerImages:
		l.Metadata = new(PreparingDockerImagesMetadata)
	case LogEventOperationDeterminingWorkspaceType:
		l.Metadata = new(DeterminingWorkspaceTypeMetadata)
	case LogEventOperationResolvingRepositories:
		l.Metadata = new(ResolvingRepositoriesMetadata)
	case LogEventOperationDeterminingWorkspaces:
		l.Metadata = new(DeterminingWorkspacesMetadata)
	case LogEventOperationCheckingCache:
		l.Metadata = new(CheckingCacheMetadata)
	case LogEventOperationExecutingTasks:
		l.Metadata = new(ExecutingTasksMetadata)
	case LogEventOperationLogFileKept:
		l.Metadata = new(LogFileKeptMetadata)
	case LogEventOperationUploadingChangesetSpecs:
		l.Metadata = new(UploadingChangesetSpecsMetadata)
	case LogEventOperationCreatingBatchSpec:
		l.Metadata = new(CreatingBatchSpecMetadata)
	case LogEventOperationApplyingBatchSpec:
		l.Metadata = new(ApplyingBatchSpecMetadata)
	case LogEventOperationBatchSpecExecution:
		l.Metadata = new(BatchSpecExecutionMetadata)
	case LogEventOperationExecutingTask:
		l.Metadata = new(ExecutingTaskMetadata)
	case LogEventOperationTaskBuildChangesetSpecs:
		l.Metadata = new(TaskBuildChangesetSpecsMetadata)
	case LogEventOperationTaskDownloadingArchive:
		l.Metadata = new(TaskDownloadingArchiveMetadata)
	case LogEventOperationTaskInitializingWorkspace:
		l.Metadata = new(TaskInitializingWorkspaceMetadata)
	case LogEventOperationTaskSkippingSteps:
		l.Metadata = new(TaskSkippingStepsMetadata)
	case LogEventOperationTaskStepSkipped:
		l.Metadata = new(TaskStepSkippedMetadata)
	case LogEventOperationTaskPreparingStep:
		l.Metadata = new(TaskPreparingStepMetadata)
	case LogEventOperationTaskStep:
		l.Metadata = new(TaskStepMetadata)
	case LogEventOperationCacheAfterStepResult:
		l.Metadata = new(CacheAfterStepResultMetadata)
	default:
		return errors.Newf("invalid event type %s", l.Operation)
	}

	wrapper := struct {
		Metadata any `json:"metadata"`
	}{
		Metadata: l.Metadata,
	}

	return json.Unmarshal(data, &wrapper)
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
	LogEventOperationCacheAfterStepResult      LogEventOperation = "CACHE_AFTER_STEP_RESULT"
)

type LogEventStatus string

const (
	LogEventStatusStarted  LogEventStatus = "STARTED"
	LogEventStatusSuccess  LogEventStatus = "SUCCESS"
	LogEventStatusFailure  LogEventStatus = "FAILURE"
	LogEventStatusProgress LogEventStatus = "PROGRESS"
)

type ParsingBatchSpecMetadata struct {
	Error string `json:"error,omitempty"`
}

type ResolvingNamespaceMetadata struct {
	NamespaceID string `json:"namespaceID,omitempty"`
}

type PreparingDockerImagesMetadata struct {
	Done  int `json:"done,omitempty"`
	Total int `json:"total,omitempty"`
}

type DeterminingWorkspaceTypeMetadata struct {
	Type string `json:"type,omitempty"`
}

type ResolvingRepositoriesMetadata struct {
	Unsupported int `json:"unsupported,omitempty"`
	Ignored     int `json:"ignored,omitempty"`
	Count       int `json:"count,omitempty"`
}

type DeterminingWorkspacesMetadata struct {
	Count int `json:"count,omitempty"`
}

type CheckingCacheMetadata struct {
	CachedSpecsFound int `json:"cachedSpecsFound,omitempty"`
	TasksToExecute   int `json:"tasksToExecute,omitempty"`
}

type JSONLinesTask struct {
	ID                     string `json:"id"`
	Repository             string `json:"repository"`
	Workspace              string `json:"workspace"`
	Steps                  []Step `json:"steps"`
	CachedStepResultsFound bool   `json:"cachedStepResultFound"`
	StartStep              int    `json:"startStep"`
}

type ExecutingTasksMetadata struct {
	Tasks   []JSONLinesTask `json:"tasks,omitempty"`
	Skipped bool            `json:"skipped,omitempty"`
	Error   string          `json:"error,omitempty"`
}

type LogFileKeptMetadata struct {
	Path string `json:"path,omitempty"`
}

type UploadingChangesetSpecsMetadata struct {
	Done  int `json:"done,omitempty"`
	Total int `json:"total,omitempty"`
	// IDs is the slice of GraphQL IDs of the created changeset specs.
	IDs []string `json:"ids,omitempty"`
}

type CreatingBatchSpecMetadata struct {
	PreviewURL string `json:"previewURL,omitempty"`
}

type ApplyingBatchSpecMetadata struct {
	BatchChangeURL string `json:"batchChangeURL,omitempty"`
}

type BatchSpecExecutionMetadata struct {
	Error string `json:"error,omitempty"`
}

type ExecutingTaskMetadata struct {
	TaskID string `json:"taskID,omitempty"`
	Error  string `json:"error,omitempty"`
}

type TaskBuildChangesetSpecsMetadata struct {
	TaskID string `json:"taskID,omitempty"`
}

type TaskDownloadingArchiveMetadata struct {
	TaskID string `json:"taskID,omitempty"`
	Error  string `json:"error,omitempty"`
}

type TaskInitializingWorkspaceMetadata struct {
	TaskID string `json:"taskID,omitempty"`
}

type TaskSkippingStepsMetadata struct {
	TaskID    string `json:"taskID,omitempty"`
	StartStep int    `json:"startStep,omitempty"`
}

type TaskStepSkippedMetadata struct {
	TaskID string `json:"taskID,omitempty"`
	Step   int    `json:"step,omitempty"`
}

type TaskPreparingStepMetadata struct {
	TaskID string `json:"taskID,omitempty"`
	Step   int    `json:"step,omitempty"`
	Error  string `json:"error,omitempty"`
}

type TaskStepMetadata struct {
	TaskID string `json:"taskID,omitempty"`
	Step   int    `json:"step,omitempty"`

	RunScript string            `json:"runScript,omitempty"`
	Env       map[string]string `json:"env,omitempty"`

	Out string `json:"out,omitempty"`

	Diff    string         `json:"diff,omitempty"`
	Outputs map[string]any `json:"outputs,omitempty"`

	ExitCode int    `json:"exitCode,omitempty"`
	Error    string `json:"error,omitempty"`
}

type CacheAfterStepResultMetadata struct {
	Key   string                    `json:"key,omitempty"`
	Value execution.AfterStepResult `json:"value,omitempty"`
}
