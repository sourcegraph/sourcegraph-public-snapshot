package batches

import (
	"encoding/json"
	"time"

	"github.com/cockroachdb/errors"
)

type LogEvent struct {
	Operation LogEventOperation `json:"operation"`

	Timestamp time.Time `json:"timestamp"`

	Status   LogEventStatus `json:"status"`
	Metadata interface{}    `json:"metadata,omitempty"`
}

func (l LogEvent) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &l); err != nil {
		return err
	}

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
	case LogEventOperationTaskCalculatingDiff:
		l.Metadata = new(TaskCalculatingDiffMetadata)
	default:
		return errors.Newf("invalid event type %s", l.Operation)
	}

	wrapper := struct {
		Metadata interface{} `json:"metadata"`
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
	LogEventOperationTaskCalculatingDiff       LogEventOperation = "TASK_CALCULATING_DIFF"
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
	Done  int
	Total int
}

type DeterminingWorkspaceTypeMetadata struct {
	Type string
}

type ResolvingRepositoriesMetadata struct {
	Unsupported int
	Ignored     int
	Count       int
}

type DeterminingWorkspacesMetadata struct {
	Count int
}

type CheckingCacheMetadata struct {
	CachedSpecsFound int
	TasksToExecute   int
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
	Tasks   []JSONLinesTask
	Skipped bool
	Error   string
}

type LogFileKeptMetadata struct {
	Path string
}

type UploadingChangesetSpecsMetadata struct {
	Done  int
	Total int
	// IDs is the slice of GraphQL IDs of the created changeset specs.
	IDs []string
}

type CreatingBatchSpecMetadata struct {
	PreviewURL string
}

type ApplyingBatchSpecMetadata struct {
	BatchChangeURL string
}

type BatchSpecExecutionMetadata struct {
	Error string
}

type ExecutingTaskMetadata struct {
	TaskID string
	Error  string
}

type TaskBuildChangesetSpecsMetadata struct {
	TaskID string
}

type TaskDownloadingArchiveMetadata struct {
	TaskID string
}

type TaskInitializingWorkspaceMetadata struct {
	TaskID string
}

type TaskSkippingStepsMetadata struct {
	TaskID    string
	StartStep int
}

type TaskStepSkippedMetadata struct {
	TaskID string
	Step   int
}

type TaskPreparingStepMetadata struct {
	TaskID string
	Step   int
	Error  string
}

type TaskStepMetadata struct {
	TaskID string
	Step   int

	RunScript string
	Env       map[string]string

	Out string

	Diff    string
	Outputs map[string]interface{}

	ExitCode int
	Error    string
}

type TaskCalculatingDiffMetadata struct {
	TaskID string
}
