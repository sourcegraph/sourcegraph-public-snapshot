package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
	"github.com/sourcegraph/src-cli/internal/batches/workspace"
)

var _ batchExecUI = &batchExecJSONLinesUI{}

type batchExecJSONLinesUI struct {
}

func (ui *batchExecJSONLinesUI) ParsingBatchSpec() {
	logOperationStart("PARSING_BATCH_SPEC", "")
}
func (ui *batchExecJSONLinesUI) ParsingBatchSpecSuccess() {
	logOperationSuccess("PARSING_BATCH_SPEC", "")
}

func (ui *batchExecJSONLinesUI) ParsingBatchSpecFailure(err error) {
	logOperationFailure("PARSING_BATCH_SPEC", err.Error())
}

func (ui *batchExecJSONLinesUI) ResolvingNamespace() {
	logOperationStart("RESOLVING_NAMESPACE", "")
}
func (ui *batchExecJSONLinesUI) ResolvingNamespaceSuccess(namespace string) {
	logOperationSuccess("RESOLVING_NAMESPACE", fmt.Sprintf("Namespace: %s", namespace))
}
func (ui *batchExecJSONLinesUI) PreparingContainerImages() {
	logOperationStart("PREPARING_DOCKER_IMAGES", "")
}
func (ui *batchExecJSONLinesUI) PreparingContainerImagesProgress(percent float64) {
	logOperationProgress("PREPARING_DOCKER_IMAGES", fmt.Sprintf("%d%% done", int(percent*100)))
}
func (ui *batchExecJSONLinesUI) PreparingContainerImagesSuccess() {
	logOperationSuccess("PREPARING_DOCKER_IMAGES", "")
}
func (ui *batchExecJSONLinesUI) DeterminingWorkspaceCreatorType() {
	logOperationStart("DETERMINING_WORKSPACE_TYPE", "")
}
func (ui *batchExecJSONLinesUI) DeterminingWorkspaceCreatorTypeSuccess(wt workspace.CreatorType) {
	switch wt {
	case workspace.CreatorTypeVolume:
		logOperationSuccess("DETERMINING_WORKSPACE_TYPE", "VOLUME")
	case workspace.CreatorTypeBind:
		logOperationSuccess("DETERMINING_WORKSPACE_TYPE", "BIND")
	}
}
func (ui *batchExecJSONLinesUI) ResolvingRepositories() {
	logOperationStart("RESOLVING_REPOSITORIES", "")
}

func (ui *batchExecJSONLinesUI) ResolvingRepositoriesDone(repos []*graphql.Repository, unsupported batches.UnsupportedRepoSet, ignored batches.IgnoredRepoSet) {
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

func (ui *batchExecJSONLinesUI) DeterminingWorkspaces() {
	logOperationStart("DETERMINING_WORKSPACES", "")
}

func (ui *batchExecJSONLinesUI) DeterminingWorkspacesSuccess(num int) {
	switch num {
	case 0:
		logOperationSuccess("DETERMINING_WORKSPACES", "No workspace found")
	case 1:
		logOperationSuccess("DETERMINING_WORKSPACES", "Found a single workspace with steps to execute")
	default:
		logOperationSuccess("DETERMINING_WORKSPACES", fmt.Sprintf("Found %d workspaces with steps to execute", num))
	}
}

func (ui *batchExecJSONLinesUI) CheckingCache() {
	logOperationStart("CHECKING_CACHE", "")
}

func (ui *batchExecJSONLinesUI) CheckingCacheSuccess(cachedSpecsFound int, tasksToExecute int) {
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

func (ui *batchExecJSONLinesUI) ExecutingTasks(verbose bool, parallelism int) func(ts []*executor.TaskStatus) {
	logOperationStart("EXECUTING_TASKS", "")

	return func(statuses []*executor.TaskStatus) {
		finishedExecution := 0
		finishedBuilding := 0
		currentlyRunning := 0
		errored := 0

		for _, ts := range statuses {
			if ts.FinishedExecution() {
				if ts.Err != nil {
					errored += 1
				}

				finishedExecution += 1
			}

			if ts.FinishedBuildingSpecs() {
				finishedBuilding += 1
			}

			if ts.IsRunning() {
				currentlyRunning += 1
			}
		}

		logOperationProgress("EXECUTING_TASKS", fmt.Sprintf("running: %d, executed: %d, built: %d, errored: %d", currentlyRunning, finishedExecution, finishedBuilding, errored))
	}
}

func (ui *batchExecJSONLinesUI) ExecutingTasksSkippingErrors(err error) {
	logOperationSuccess("EXECUTING_TASKS", fmt.Sprintf("Error: %s. Skipping errors because -skip-errors was used.", err))
}
func (ui *batchExecJSONLinesUI) ExecutingTasksSuccess() {

	logOperationSuccess("EXECUTING_TASKS", "")
}

func (ui *batchExecJSONLinesUI) LogFilesKept(files []string) {
	for _, file := range files {
		logOperationSuccess("LOG_FILE_KEPT", file)
	}
}

func (ui *batchExecJSONLinesUI) NoChangesetSpecs() {
	logOperationSuccess("UPLOADING_CHANGESET_SPECS", "No changeset specs created")
}

func (ui *batchExecJSONLinesUI) UploadingChangesetSpecs(num int) {
	var label string
	if num == 1 {
		label = "Sending 1 changeset spec"
	} else {
		label = fmt.Sprintf("Sending %d changeset specs", num)
	}

	logOperationStart("UPLOADING_CHANGESET_SPECS", label)
}

func (ui *batchExecJSONLinesUI) UploadingChangesetSpecsProgress(done, total int) {
	logOperationProgress("UPLOADING_CHANGESET_SPECS", fmt.Sprintf("Uploaded %d out of %d", done, total))
}
func (ui *batchExecJSONLinesUI) UploadingChangesetSpecsSuccess() {
	logOperationSuccess("UPLOADING_CHANGESET_SPECS", "")
}

func (ui *batchExecJSONLinesUI) CreatingBatchSpec() {
	logOperationStart("CREATING_BATCH_SPEC", "")
}

func (ui *batchExecJSONLinesUI) CreatingBatchSpecSuccess() {
}

func (ui *batchExecJSONLinesUI) CreatingBatchSpecError(err error) error {
	return err
}

func (ui *batchExecJSONLinesUI) PreviewBatchSpec(batchSpecURL string) {
	logOperationSuccess("CREATING_BATCH_SPEC", batchSpecURL)
}

func (ui *batchExecJSONLinesUI) ApplyingBatchSpec() {
	logOperationStart("APPLYING_BATCH_SPEC", "")
}

func (ui *batchExecJSONLinesUI) ApplyingBatchSpecSuccess(batchChangeURL string) {
	logOperationSuccess("APPLYING_BATCH_SPEC", batchChangeURL)
}

func (ui *batchExecJSONLinesUI) ExecutionError(err error) {
	logOperationFailure("BATCH_SPEC_EXECUTION", err.Error())
}

type batchesLogEvent struct {
	Operation string `json:"operation"` // "PREPARING_DOCKER_IMAGES"

	Timestamp time.Time `json:"timestamp"`

	Status  string `json:"status"`            // "STARTED", "PROGRESS", "SUCCESS", "FAILURE"
	Message string `json:"message,omitempty"` // "70% done"
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
	json.NewEncoder(os.Stdout).Encode(e)
}
