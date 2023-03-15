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
	case LogEventOperationDockerWatchDog:
		l.Metadata = new(DockerWatchDogMetadata)
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
	LogEventOperationParsingBatchSpec         LogEventOperation = "PARSING_BATCH_SPEC"
	LogEventOperationResolvingNamespace       LogEventOperation = "RESOLVING_NAMESPACE"
	LogEventOperationPreparingDockerImages    LogEventOperation = "PREPARING_DOCKER_IMAGES"
	LogEventOperationDeterminingWorkspaceType LogEventOperation = "DETERMINING_WORKSPACE_TYPE"
	LogEventOperationDeterminingWorkspaces    LogEventOperation = "DETERMINING_WORKSPACES"
	LogEventOperationCheckingCache            LogEventOperation = "CHECKING_CACHE"
	LogEventOperationExecutingTasks           LogEventOperation = "EXECUTING_TASKS"
	LogEventOperationLogFileKept              LogEventOperation = "LOG_FILE_KEPT"
	LogEventOperationUploadingChangesetSpecs  LogEventOperation = "UPLOADING_CHANGESET_SPECS"
	LogEventOperationCreatingBatchSpec        LogEventOperation = "CREATING_BATCH_SPEC"
	LogEventOperationApplyingBatchSpec        LogEventOperation = "APPLYING_BATCH_SPEC"
	LogEventOperationBatchSpecExecution       LogEventOperation = "BATCH_SPEC_EXECUTION"
	LogEventOperationExecutingTask            LogEventOperation = "EXECUTING_TASK"
	LogEventOperationTaskBuildChangesetSpecs  LogEventOperation = "TASK_BUILD_CHANGESET_SPECS"
	LogEventOperationTaskSkippingSteps        LogEventOperation = "TASK_SKIPPING_STEPS"
	LogEventOperationTaskStepSkipped          LogEventOperation = "TASK_STEP_SKIPPED"
	LogEventOperationTaskPreparingStep        LogEventOperation = "TASK_PREPARING_STEP"
	LogEventOperationTaskStep                 LogEventOperation = "TASK_STEP"
	LogEventOperationCacheAfterStepResult     LogEventOperation = "CACHE_AFTER_STEP_RESULT"
	LogEventOperationDockerWatchDog           LogEventOperation = "DOCKER_WATCH_DOG"
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

type DeterminingWorkspacesMetadata struct {
	Unsupported    int `json:"unsupported,omitempty"`
	Ignored        int `json:"ignored,omitempty"`
	RepoCount      int `json:"repoCount,omitempty"`
	WorkspaceCount int `json:"workspaceCount,omitempty"`
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
	Version int
	TaskID  string
	Step    int

	RunScript string
	Env       map[string]string

	Out string

	Diff    []byte
	Outputs map[string]any

	ExitCode int
	Error    string
}

func (m TaskStepMetadata) MarshalJSON() ([]byte, error) {
	if m.Version == 2 {
		return json.Marshal(v2TaskStepMetadata{
			Version:   2,
			TaskID:    m.TaskID,
			Step:      m.Step,
			RunScript: m.RunScript,
			Env:       m.Env,
			Out:       m.Out,
			Diff:      m.Diff,
			Outputs:   m.Outputs,
			ExitCode:  m.ExitCode,
			Error:     m.Error,
		})
	}
	return json.Marshal(v1TaskStepMetadata{
		TaskID:    m.TaskID,
		Step:      m.Step,
		RunScript: m.RunScript,
		Env:       m.Env,
		Out:       m.Out,
		Diff:      string(m.Diff),
		Outputs:   m.Outputs,
		ExitCode:  m.ExitCode,
		Error:     m.Error,
	})
}

func (m *TaskStepMetadata) UnmarshalJSON(data []byte) error {
	var version versionTaskStepMetadata
	if err := json.Unmarshal(data, &version); err != nil {
		return err
	}
	if version.Version == 2 {
		var v2 v2TaskStepMetadata
		if err := json.Unmarshal(data, &v2); err != nil {
			return err
		}
		m.Version = v2.Version
		m.TaskID = v2.TaskID
		m.Step = v2.Step
		m.RunScript = v2.RunScript
		m.Env = v2.Env
		m.Out = v2.Out
		m.Diff = v2.Diff
		m.Outputs = v2.Outputs
		m.ExitCode = v2.ExitCode
		m.Error = v2.Error
		return nil
	}
	var v1 v1TaskStepMetadata
	if err := json.Unmarshal(data, &v1); err != nil {
		return errors.Wrap(err, string(data))
	}
	m.TaskID = v1.TaskID
	m.Step = v1.Step
	m.RunScript = v1.RunScript
	m.Env = v1.Env
	m.Out = v1.Out
	m.Diff = []byte(v1.Diff)
	m.Outputs = v1.Outputs
	m.ExitCode = v1.ExitCode
	m.Error = v1.Error
	return nil
}

type versionTaskStepMetadata struct {
	Version int `json:"version,omitempty"`
}

type v2TaskStepMetadata struct {
	Version   int               `json:"version,omitempty"`
	TaskID    string            `json:"taskID,omitempty"`
	Step      int               `json:"step,omitempty"`
	RunScript string            `json:"runScript,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Out       string            `json:"out,omitempty"`
	Diff      []byte            `json:"diff,omitempty"`
	Outputs   map[string]any    `json:"outputs,omitempty"`
	ExitCode  int               `json:"exitCode,omitempty"`
	Error     string            `json:"error,omitempty"`
}

type v1TaskStepMetadata struct {
	TaskID    string            `json:"taskID,omitempty"`
	Step      int               `json:"step,omitempty"`
	RunScript string            `json:"runScript,omitempty"`
	Env       map[string]string `json:"env,omitempty"`
	Out       string            `json:"out,omitempty"`
	Diff      string            `json:"diff,omitempty"`
	Outputs   map[string]any    `json:"outputs,omitempty"`
	ExitCode  int               `json:"exitCode,omitempty"`
	Error     string            `json:"error,omitempty"`
}

type CacheAfterStepResultMetadata struct {
	Key   string                    `json:"key,omitempty"`
	Value execution.AfterStepResult `json:"value,omitempty"`
}

type DockerWatchDogMetadata struct {
	Error string `json:"error,omitempty"`
}
