package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
)

var _ ExecUI = &JSONLines{}

type JSONLines struct{}

func (ui *JSONLines) ParsingBatchSpec() {
	logOperationStart("PARSING_BATCH_SPEC", "")
}
func (ui *JSONLines) ParsingBatchSpecSuccess() {
	logOperationSuccess("PARSING_BATCH_SPEC", "")
}

func (ui *JSONLines) ParsingBatchSpecFailure(err error) {
	logOperationFailure("PARSING_BATCH_SPEC", err.Error())
}

func (ui *JSONLines) ResolvingNamespace() {
	logOperationStart("RESOLVING_NAMESPACE", "")
}
func (ui *JSONLines) ResolvingNamespaceSuccess(namespace string) {
	logOperationSuccess("RESOLVING_NAMESPACE", fmt.Sprintf("Namespace: %s", namespace))
}
func (ui *JSONLines) PreparingContainerImages() {
	logOperationStart("PREPARING_DOCKER_IMAGES", "")
}
func (ui *JSONLines) PreparingContainerImagesProgress(percent float64) {
	logOperationProgress("PREPARING_DOCKER_IMAGES", fmt.Sprintf("%d%% done", int(percent*100)))
}
func (ui *JSONLines) PreparingContainerImagesSuccess() {
	logOperationSuccess("PREPARING_DOCKER_IMAGES", "")
}
func (ui *JSONLines) DeterminingWorkspaceCreatorType() {
	logOperationStart("DETERMINING_WORKSPACE_TYPE", "")
}
func (ui *JSONLines) DeterminingWorkspaceCreatorTypeSuccess(wt workspace.CreatorType) {
	switch wt {
	case workspace.CreatorTypeVolume:
		logOperationSuccess("DETERMINING_WORKSPACE_TYPE", "VOLUME")
	case workspace.CreatorTypeBind:
		logOperationSuccess("DETERMINING_WORKSPACE_TYPE", "BIND")
	}
}
func (ui *JSONLines) ResolvingRepositories() {
	logOperationStart("RESOLVING_REPOSITORIES", "")
}

func (ui *JSONLines) ResolvingRepositoriesDone(repos []*graphql.Repository, unsupported batches.UnsupportedRepoSet, ignored batches.IgnoredRepoSet) {
	if unsupported != nil && len(unsupported) != 0 {
		logOperationSuccess("RESOLVING_REPOSITORIES", fmt.Sprintf("%d unsupported repositories", len(unsupported)))
	} else if ignored != nil && len(ignored) != 0 {
		logOperationSuccess("RESOLVING_REPOSITORIES", fmt.Sprintf("%d ignored repositories", len(ignored)))
	} else {
		switch len(repos) {
		case 0:
			logOperationSuccess("RESOLVING_REPOSITORIES", "No repositories resolved")
		case 1:
			logOperationSuccess("RESOLVING_REPOSITORIES", "Resolved 1 repository")
		default:
			logOperationSuccess("RESOLVING_REPOSITORIES", fmt.Sprintf("Resolved %d repositories", len(repos)))
		}
	}
}

func (ui *JSONLines) DeterminingWorkspaces() {
	logOperationStart("DETERMINING_WORKSPACES", "")
}

func (ui *JSONLines) DeterminingWorkspacesSuccess(num int) {
	switch num {
	case 0:
		logOperationSuccess("DETERMINING_WORKSPACES", "No workspace found")
	case 1:
		logOperationSuccess("DETERMINING_WORKSPACES", "Found a single workspace with steps to execute")
	default:
		logOperationSuccess("DETERMINING_WORKSPACES", fmt.Sprintf("Found %d workspaces with steps to execute", num))
	}
}

func (ui *JSONLines) CheckingCache() {
	logOperationStart("CHECKING_CACHE", "")
}

func (ui *JSONLines) CheckingCacheSuccess(cachedSpecsFound int, tasksToExecute int) {
	var specsFoundMessage string
	if cachedSpecsFound == 1 {
		specsFoundMessage = "Found 1 cached changeset spec"
	} else {
		specsFoundMessage = fmt.Sprintf("Found %d cached changeset specs", cachedSpecsFound)
	}
	switch tasksToExecute {
	case 0:
		logOperationSuccess("CHECKING_CACHE", fmt.Sprintf("%s; no tasks need to be executed", specsFoundMessage))
	case 1:
		logOperationSuccess("CHECKING_CACHE", fmt.Sprintf("%s; %d task needs to be executed", specsFoundMessage, tasksToExecute))
	default:
		logOperationSuccess("CHECKING_CACHE", fmt.Sprintf("%s; %d tasks need to be executed", specsFoundMessage, tasksToExecute))
	}
}

func (ui *JSONLines) ExecutingTasks(verbose bool, parallelism int) executor.TaskExecutionUI {
	return &taskExecutionJSONLines{verbose: verbose, parallelism: parallelism}
}

func (ui *JSONLines) ExecutingTasksSkippingErrors(err error) {
	logOperationSuccess("EXECUTING_TASKS", fmt.Sprintf("Error: %s. Skipping errors because -skip-errors was used.", err))
}

func (ui *JSONLines) LogFilesKept(files []string) {
	for _, file := range files {
		logOperationSuccess("LOG_FILE_KEPT", file)
	}
}

func (ui *JSONLines) NoChangesetSpecs() {
	logOperationSuccess("UPLOADING_CHANGESET_SPECS", "No changeset specs created")
}

func (ui *JSONLines) UploadingChangesetSpecs(num int) {
	var label string
	if num == 1 {
		label = "Sending 1 changeset spec"
	} else {
		label = fmt.Sprintf("Sending %d changeset specs", num)
	}

	logOperationStart("UPLOADING_CHANGESET_SPECS", label)
}

func (ui *JSONLines) UploadingChangesetSpecsProgress(done, total int) {
	logOperationProgress("UPLOADING_CHANGESET_SPECS", fmt.Sprintf("Uploaded %d out of %d", done, total))
}
func (ui *JSONLines) UploadingChangesetSpecsSuccess() {
	logOperationSuccess("UPLOADING_CHANGESET_SPECS", "")
}

func (ui *JSONLines) CreatingBatchSpec() {
	logOperationStart("CREATING_BATCH_SPEC", "")
}

func (ui *JSONLines) CreatingBatchSpecSuccess() {
}

func (ui *JSONLines) CreatingBatchSpecError(err error) error {
	return err
}

func (ui *JSONLines) PreviewBatchSpec(batchSpecURL string) {
	logOperationSuccess("CREATING_BATCH_SPEC", batchSpecURL)
}

func (ui *JSONLines) ApplyingBatchSpec() {
	logOperationStart("APPLYING_BATCH_SPEC", "")
}

func (ui *JSONLines) ApplyingBatchSpecSuccess(batchChangeURL string) {
	logOperationSuccess("APPLYING_BATCH_SPEC", batchChangeURL)
}

func (ui *JSONLines) ExecutionError(err error) {
	logOperationFailure("BATCH_SPEC_EXECUTION", err.Error())
}

type batchesLogEvent struct {
	Operation string `json:"operation"` // "PREPARING_DOCKER_IMAGES"

	Timestamp time.Time `json:"timestamp"`

	Status   string      `json:"status"`            // "STARTED", "PROGRESS", "SUCCESS", "FAILURE"
	Message  string      `json:"message,omitempty"` // "70% done"
	Metadata interface{} `json:"metadata,omitempty"`
}

func logOperationStart(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "STARTED", Message: msg})
}

func logOperationSuccess(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "SUCCESS", Message: msg})
}

func logOperationFailure(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "FAILURE", Message: msg})
}

func logOperationProgress(op, msg string) {
	logEvent(batchesLogEvent{Operation: op, Status: "PROGRESS", Message: msg})
}

func logEvent(e batchesLogEvent) {
	e.Timestamp = time.Now().UTC().Truncate(time.Millisecond)
	err := json.NewEncoder(os.Stdout).Encode(e)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// TODO: Until we've figured out what exactly we want to expose, we create
// these smaller UI-specific structs.
type jsonLinesTask struct {
	Repository             string
	Workspace              string
	Steps                  []batcheslib.Step
	CachedStepResultsFound bool
	StartStep              int
}

type taskExecutionJSONLines struct {
	verbose     bool
	parallelism int

	linesTasks map[*executor.Task]jsonLinesTask
}

func (ui *taskExecutionJSONLines) Start(tasks []*executor.Task) {
	ui.linesTasks = make(map[*executor.Task]jsonLinesTask, len(tasks))
	linesTasks := []jsonLinesTask{}
	for _, t := range tasks {
		linesTask := jsonLinesTask{
			Repository:             t.Repository.Name,
			Workspace:              t.Path,
			Steps:                  t.Steps,
			CachedStepResultsFound: t.CachedResultFound,
			StartStep:              t.CachedResult.StepIndex,
		}
		ui.linesTasks[t] = linesTask
		linesTasks = append(linesTasks, linesTask)
	}

	logEvent(batchesLogEvent{Operation: "EXECUTING_TASKS", Status: "STARTED", Metadata: map[string]interface{}{
		"tasks": linesTasks,
	}})
}

func (ui *taskExecutionJSONLines) Success() {
	logEvent(batchesLogEvent{Operation: "EXECUTING_TASKS", Status: "SUCCESS"})
}

func (ui *taskExecutionJSONLines) TaskStarted(task *executor.Task) {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}
	logEvent(batchesLogEvent{Operation: "EXECUTING_TASK", Status: "STARTED", Metadata: map[string]interface{}{
		"task": lt,
	}})
}

func (ui *taskExecutionJSONLines) TaskFinished(task *executor.Task, err error) {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}
	if err != nil {
		logEvent(batchesLogEvent{Operation: "EXECUTING_TASK", Status: "FAILURE", Metadata: map[string]interface{}{
			"task":  lt,
			"error": err,
		}})
		return
	}

	logEvent(batchesLogEvent{Operation: "EXECUTING_TASK", Status: "SUCCESS", Metadata: map[string]interface{}{
		"task": lt,
	}})
}

func (ui *taskExecutionJSONLines) TaskChangesetSpecsBuilt(task *executor.Task, specs []*batcheslib.ChangesetSpec) {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}
	logEvent(batchesLogEvent{Operation: "BUILDING_TASK_CHANGESET_SPECS", Status: "SUCCESS", Metadata: map[string]interface{}{
		"task": lt,
	}})
}

func (ui *taskExecutionJSONLines) TaskCurrentlyExecuting(task *executor.Task, message string) {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}

	logEvent(batchesLogEvent{
		Operation: "EXECUTING_TASK",
		Status:    "PROGRESS",
		Message:   message,
		Metadata: map[string]interface{}{
			"task": lt,
		},
	})
}

const stepFlushDuration = 500 * time.Millisecond

func (ui *taskExecutionJSONLines) StepStdoutWriter(ctx context.Context, task *executor.Task, step int) io.WriteCloser {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}

	sink := func(data string) {
		logEvent(batchesLogEvent{
			Operation: "STEP",
			Status:    "PROGRESS",
			Message:   data,
			Metadata: map[string]interface{}{
				"task":        lt,
				"step":        step,
				"output_type": "stdout",
			},
		})
	}
	return NewIntervalWriter(ctx, stepFlushDuration, sink)
}

func (ui *taskExecutionJSONLines) StepStderrWriter(ctx context.Context, task *executor.Task, step int) io.WriteCloser {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}

	sink := func(data string) {
		logEvent(batchesLogEvent{
			Operation: "STEP",
			Status:    "PROGRESS",
			Message:   data,
			Metadata: map[string]interface{}{
				"task":        lt,
				"step":        step,
				"output_type": "stderr",
			},
		})
	}

	return NewIntervalWriter(ctx, stepFlushDuration, sink)
}
