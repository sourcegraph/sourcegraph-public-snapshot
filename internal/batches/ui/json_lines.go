package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/dineshappavoo/basex"

	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
)

type BatchesLogEvent struct {
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

var _ ExecUI = &JSONLines{}

type JSONLines struct{}

func (ui *JSONLines) ParsingBatchSpec() {
	logOperationStart(LogEventOperationParsingBatchSpec, nil)
}
func (ui *JSONLines) ParsingBatchSpecSuccess() {
	logOperationSuccess(LogEventOperationParsingBatchSpec, nil)
}
func (ui *JSONLines) ParsingBatchSpecFailure(err error) {
	logOperationFailure(LogEventOperationParsingBatchSpec, map[string]interface{}{"error": err.Error()})
}

func (ui *JSONLines) ResolvingNamespace() {
	logOperationStart(LogEventOperationResolvingNamespace, nil)
}
func (ui *JSONLines) ResolvingNamespaceSuccess(namespace string) {
	logOperationSuccess(LogEventOperationResolvingNamespace, map[string]interface{}{"namespaceID": namespace})
}

func (ui *JSONLines) PreparingContainerImages() {
	logOperationStart(LogEventOperationPreparingDockerImages, nil)
}
func (ui *JSONLines) PreparingContainerImagesProgress(done, total int) {
	logOperationProgress(LogEventOperationPreparingDockerImages, map[string]interface{}{"done": done, "total": total})
}
func (ui *JSONLines) PreparingContainerImagesSuccess() {
	logOperationSuccess(LogEventOperationPreparingDockerImages, nil)
}

func (ui *JSONLines) DeterminingWorkspaceCreatorType() {
	logOperationStart(LogEventOperationDeterminingWorkspaceType, nil)
}
func (ui *JSONLines) DeterminingWorkspaceCreatorTypeSuccess(wt workspace.CreatorType) {
	switch wt {
	case workspace.CreatorTypeVolume:
		logOperationSuccess(LogEventOperationDeterminingWorkspaceType, map[string]interface{}{"type": "VOLUME"})
	case workspace.CreatorTypeBind:
		logOperationSuccess(LogEventOperationDeterminingWorkspaceType, map[string]interface{}{"type": "BIND"})
	}
}

func (ui *JSONLines) ResolvingRepositories() {
	logOperationStart(LogEventOperationResolvingRepositories, nil)
}
func (ui *JSONLines) ResolvingRepositoriesDone(repos []*graphql.Repository, unsupported batches.UnsupportedRepoSet, ignored batches.IgnoredRepoSet) {
	if unsupported != nil && len(unsupported) != 0 {
		logOperationSuccess(LogEventOperationResolvingRepositories, map[string]interface{}{"unsupported": len(unsupported)})
	} else if ignored != nil && len(ignored) != 0 {
		logOperationSuccess(LogEventOperationResolvingRepositories, map[string]interface{}{"ignored": len(ignored)})
	} else {
		logOperationSuccess(LogEventOperationResolvingRepositories, map[string]interface{}{"count": len(repos)})
	}
}

func (ui *JSONLines) DeterminingWorkspaces() {
	logOperationStart(LogEventOperationDeterminingWorkspaces, nil)
}
func (ui *JSONLines) DeterminingWorkspacesSuccess(num int) {
	metadata := map[string]interface{}{
		"count": num,
	}
	logOperationSuccess(LogEventOperationDeterminingWorkspaces, metadata)
}

func (ui *JSONLines) CheckingCache() {
	logOperationStart(LogEventOperationCheckingCache, nil)
}
func (ui *JSONLines) CheckingCacheSuccess(cachedSpecsFound int, tasksToExecute int) {
	metadata := map[string]interface{}{
		"cachedSpecsFound": cachedSpecsFound,
		"tasksToExecute":   tasksToExecute,
	}
	logOperationSuccess(LogEventOperationCheckingCache, metadata)
}

func (ui *JSONLines) ExecutingTasks(verbose bool, parallelism int) executor.TaskExecutionUI {
	return &taskExecutionJSONLines{verbose: verbose, parallelism: parallelism}
}

func (ui *JSONLines) ExecutingTasksSkippingErrors(err error) {
	logOperationSuccess(LogEventOperationExecutingTasks, map[string]interface{}{"skipped": true, "error": err.Error()})
}

func (ui *JSONLines) LogFilesKept(files []string) {
	for _, file := range files {
		logOperationSuccess(LogEventOperationLogFileKept, map[string]interface{}{"path": file})
	}
}

func (ui *JSONLines) NoChangesetSpecs() {
	ui.UploadingChangesetSpecsSuccess([]graphql.ChangesetSpecID{})
}

func (ui *JSONLines) UploadingChangesetSpecs(num int) {
	logOperationStart(LogEventOperationUploadingChangesetSpecs, map[string]interface{}{
		"total": num,
	})
}

func (ui *JSONLines) UploadingChangesetSpecsProgress(done, total int) {
	logOperationProgress(LogEventOperationUploadingChangesetSpecs, map[string]interface{}{
		"done":  done,
		"total": total,
	})
}

func (ui *JSONLines) UploadingChangesetSpecsSuccess(ids []graphql.ChangesetSpecID) {
	logOperationSuccess(LogEventOperationUploadingChangesetSpecs, map[string]interface{}{
		"ids": ids,
	})
}

func (ui *JSONLines) CreatingBatchSpec() {
	logOperationStart(LogEventOperationCreatingBatchSpec, nil)
}

func (ui *JSONLines) CreatingBatchSpecSuccess() {
}

func (ui *JSONLines) CreatingBatchSpecError(err error) error {
	return err
}

func (ui *JSONLines) PreviewBatchSpec(batchSpecURL string) {
	logOperationSuccess(LogEventOperationCreatingBatchSpec, map[string]interface{}{"batchSpecURL": batchSpecURL})
}

func (ui *JSONLines) ApplyingBatchSpec() {
	logOperationStart(LogEventOperationApplyingBatchSpec, nil)
}

func (ui *JSONLines) ApplyingBatchSpecSuccess(batchChangeURL string) {
	logOperationSuccess(LogEventOperationApplyingBatchSpec, map[string]interface{}{"batchChangeURL": batchChangeURL})
}

func (ui *JSONLines) ExecutionError(err error) {
	logOperationFailure(LogEventOperationBatchSpecExecution, map[string]interface{}{"error": err.Error()})
}

func logOperationStart(op LogEventOperation, metadata map[string]interface{}) {
	logEvent(BatchesLogEvent{Operation: op, Status: LogEventStatusStarted, Metadata: metadata})
}

func logOperationSuccess(op LogEventOperation, metadata map[string]interface{}) {
	logEvent(BatchesLogEvent{Operation: op, Status: LogEventStatusSuccess, Metadata: metadata})
}

func logOperationFailure(op LogEventOperation, metadata map[string]interface{}) {
	logEvent(BatchesLogEvent{Operation: op, Status: LogEventStatusFailure, Metadata: metadata})
}

func logOperationProgress(op LogEventOperation, metadata map[string]interface{}) {
	logEvent(BatchesLogEvent{Operation: op, Status: LogEventStatusProgress, Metadata: metadata})
}

func logEvent(e BatchesLogEvent) {
	e.Timestamp = time.Now().UTC().Truncate(time.Millisecond)
	err := json.NewEncoder(os.Stdout).Encode(e)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}

// TODO: Until we've figured out what exactly we want to expose, we create
// these smaller UI-specific structs.
type jsonLinesTask struct {
	ID                     string            `json:"id"`
	Repository             string            `json:"repository"`
	Workspace              string            `json:"workspace"`
	Steps                  []batcheslib.Step `json:"steps"`
	CachedStepResultsFound bool              `json:"cachedStepResultFound"`
	StartStep              int               `json:"startStep"`
}

type taskExecutionJSONLines struct {
	verbose     bool
	parallelism int

	linesTasks map[*executor.Task]jsonLinesTask
}

// seededRand is used in randomID() to generate a "random" number.
var seededRand = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

// randomID generates a random ID to be used for identifiers in tasks.
func randomID() (string, error) {
	return basex.Encode(strconv.Itoa(seededRand.Int()))
}

func (ui *taskExecutionJSONLines) Start(tasks []*executor.Task) {
	ui.linesTasks = make(map[*executor.Task]jsonLinesTask, len(tasks))
	linesTasks := []jsonLinesTask{}
	for _, t := range tasks {
		id, err := randomID()
		if err != nil {
			panic(err)
		}
		linesTask := jsonLinesTask{
			ID:                     id,
			Repository:             t.Repository.Name,
			Workspace:              t.Path,
			Steps:                  t.Steps,
			CachedStepResultsFound: t.CachedResultFound,
			StartStep:              t.CachedResult.StepIndex,
		}
		ui.linesTasks[t] = linesTask
		linesTasks = append(linesTasks, linesTask)
	}

	logOperationStart(LogEventOperationExecutingTasks, map[string]interface{}{
		"tasks": linesTasks,
	})
}
func (ui *taskExecutionJSONLines) Success() {
	logOperationSuccess(LogEventOperationExecutingTasks, nil)
}

func (ui *taskExecutionJSONLines) TaskStarted(task *executor.Task) {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}

	logOperationStart(LogEventOperationExecutingTask, map[string]interface{}{
		"taskID": lt.ID,
	})
}

func (ui *taskExecutionJSONLines) TaskFinished(task *executor.Task, err error) {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}

	if err != nil {
		logOperationFailure(LogEventOperationExecutingTask, map[string]interface{}{
			"taskID": lt.ID,
			"error":  err,
		})
		return
	}

	logOperationSuccess(LogEventOperationExecutingTask, map[string]interface{}{
		"taskID": lt.ID,
	})
}

func (ui *taskExecutionJSONLines) TaskChangesetSpecsBuilt(task *executor.Task, specs []*batcheslib.ChangesetSpec) {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}
	logOperationSuccess(LogEventOperationTaskBuildChangesetSpecs, map[string]interface{}{
		"taskID": lt.ID,
	})
}

func (ui *taskExecutionJSONLines) StepsExecutionUI(task *executor.Task) executor.StepsExecutionUI {
	lt, ok := ui.linesTasks[task]
	if !ok {
		panic("unknown task started")
	}

	return &stepsExecutionJSONLines{linesTask: &lt}
}

type stepsExecutionJSONLines struct {
	linesTask *jsonLinesTask
}

const stepFlushDuration = 500 * time.Millisecond

func (ui *stepsExecutionJSONLines) ArchiveDownloadStarted() {
	logOperationStart(LogEventOperationTaskDownloadingArchive, map[string]interface{}{"taskID": ui.linesTask.ID})
}

func (ui *stepsExecutionJSONLines) ArchiveDownloadFinished() {
	logOperationSuccess(LogEventOperationTaskDownloadingArchive, map[string]interface{}{"taskID": ui.linesTask.ID})
}
func (ui *stepsExecutionJSONLines) WorkspaceInitializationStarted() {
	logOperationStart(LogEventOperationTaskInitializingWorkspace, map[string]interface{}{"taskID": ui.linesTask.ID})
}
func (ui *stepsExecutionJSONLines) WorkspaceInitializationFinished() {
	logOperationSuccess(LogEventOperationTaskInitializingWorkspace, map[string]interface{}{"taskID": ui.linesTask.ID})
}

func (ui *stepsExecutionJSONLines) SkippingStepsUpto(startStep int) {
	logOperationProgress(LogEventOperationTaskSkippingSteps, map[string]interface{}{"taskID": ui.linesTask.ID, "startStep": startStep})
}

func (ui *stepsExecutionJSONLines) StepSkipped(step int) {
	logOperationProgress(LogEventOperationTaskStepSkipped, map[string]interface{}{"taskID": ui.linesTask.ID, "step": step})
}

func (ui *stepsExecutionJSONLines) StepPreparing(step int) {
	logOperationProgress(LogEventOperationTaskPreparingStep, map[string]interface{}{"taskID": ui.linesTask.ID, "step": step})
}

func (ui *stepsExecutionJSONLines) StepStarted(step int, runScript string) {
	logOperationStart(LogEventOperationTaskStep, map[string]interface{}{"taskID": ui.linesTask.ID, "step": step, "runScript": runScript})
}

func (ui *stepsExecutionJSONLines) StepStdoutWriter(ctx context.Context, task *executor.Task, step int) io.WriteCloser {
	sink := func(data string) {
		logOperationProgress(
			LogEventOperationTaskStep,
			map[string]interface{}{
				"taskID":      ui.linesTask.ID,
				"step":        step,
				"out":         data,
				"output_type": "stdout",
			},
		)
	}
	return NewIntervalWriter(ctx, stepFlushDuration, sink)
}

func (ui *stepsExecutionJSONLines) StepStderrWriter(ctx context.Context, task *executor.Task, step int) io.WriteCloser {
	sink := func(data string) {
		logOperationProgress(
			LogEventOperationTaskStep,
			map[string]interface{}{
				"taskID":      ui.linesTask.ID,
				"step":        step,
				"out":         data,
				"output_type": "stderr",
			},
		)
	}

	return NewIntervalWriter(ctx, stepFlushDuration, sink)
}

func (ui *stepsExecutionJSONLines) StepFinished(step int, diff []byte, changes *git.Changes, outputs map[string]interface{}) {
	logOperationSuccess(
		LogEventOperationTaskStep,
		map[string]interface{}{
			"task":    ui.linesTask,
			"step":    step,
			"diff":    string(diff),
			"changes": changes,
			"outputs": outputs,
		},
	)
}

func (ui *stepsExecutionJSONLines) CalculatingDiffStarted() {
	logOperationStart(LogEventOperationTaskCalculatingDiff, map[string]interface{}{"taskID": ui.linesTask.ID})
}
func (ui *stepsExecutionJSONLines) CalculatingDiffFinished() {
	logOperationSuccess(LogEventOperationTaskCalculatingDiff, map[string]interface{}{"taskID": ui.linesTask.ID})
}
