package ui

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/src-cli/internal/batches/executor"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/batches/git"
	"github.com/sourcegraph/sourcegraph/lib/output"
)

type taskStatus struct {
	displayName string

	startedAt          time.Time
	finishedAt         time.Time
	currentlyExecuting string

	// err is set if executing the Task lead to an error.
	err error
}

func (ts *taskStatus) FinishedExecution() bool {
	return !ts.startedAt.IsZero() && !ts.finishedAt.IsZero()
}

func (ts *taskStatus) ExecutionTime() time.Duration {
	return ts.finishedAt.Sub(ts.startedAt).Truncate(time.Millisecond)
}

func (ts *taskStatus) String() string {
	var statusText string

	if ts.FinishedExecution() {
		if ts.err != nil {
			if texter, ok := ts.err.(statusTexter); ok {
				statusText = texter.StatusText()
			} else {
				statusText = ts.err.Error()
			}
		} else {
			statusText = "Done!"
		}
	} else {
		if ts.currentlyExecuting != "" {
			lines := strings.Split(ts.currentlyExecuting, "\n")
			escapedLine := strings.ReplaceAll(lines[0], "%", "%%")
			if len(lines) > 1 {
				statusText = fmt.Sprintf("%s ...", escapedLine)
			} else {
				statusText = escapedLine
			}
		} else {
			statusText = "..."
		}
	}

	return statusText
}

type clock func() time.Time

var defaultClock = time.Now

func newTaskExecTUI(out *output.Output, verbose bool, numParallelism int) *taskExecTUI {
	return &taskExecTUI{
		out:            out,
		verbose:        verbose,
		numParallelism: numParallelism,

		clock: defaultClock,

		statuses:   map[*executor.Task]*taskStatus{},
		statusBars: map[int]*taskStatus{},
	}
}

type taskExecTUI struct {
	// Used in tests only
	forceNoSpinner bool

	out *output.Output

	verbose bool

	progress      output.ProgressWithStatusBars
	numStatusBars int

	maxRepoName    int
	numParallelism int

	mu    sync.Mutex
	clock clock

	statuses   map[*executor.Task]*taskStatus
	statusBars map[int]*taskStatus

	finished int
	errored  int
}

var _ executor.TaskExecutionUI = &taskExecTUI{}

func (ui *taskExecTUI) Start(tasks []*executor.Task) {
	for _, t := range tasks {
		status := &taskStatus{}
		if t.Path != "" {
			status.displayName = t.Repository.Name + ":" + t.Path
		} else {
			status.displayName = t.Repository.Name
		}

		if len(status.displayName) > ui.maxRepoName {
			ui.maxRepoName = len(status.displayName)
		}

		ui.statuses[t] = status
	}

	ui.numStatusBars = ui.numParallelism
	if len(tasks) < ui.numStatusBars {
		ui.numStatusBars = len(tasks)
	}

	statusBars := make([]*output.StatusBar, 0, ui.numStatusBars)
	for i := 0; i < ui.numStatusBars; i++ {
		statusBars = append(statusBars, output.NewStatusBar())
	}

	progressBars := []output.ProgressBar{
		{
			Label: fmt.Sprintf("Executing... (0/%d, 0 errored)", len(tasks)),
			Max:   float64(len(tasks)),
		},
	}

	opts := output.DefaultProgressTTYOpts.WithNoSpinner(ui.forceNoSpinner)
	ui.progress = ui.out.ProgressWithStatusBars(progressBars, statusBars, opts)
}

func (ui *taskExecTUI) Success() {
	ui.progress.Complete()
}
func (ui *taskExecTUI) Failed(err error) {
	// noop right now
}

func (ui *taskExecTUI) useFreeStatusBar(ts *taskStatus) (bar int, found bool) {
	for i := 0; i < ui.numStatusBars; i++ {
		if _, ok := ui.statusBars[i]; !ok {
			ui.statusBars[i] = ts
			bar = i
			found = true
			return bar, found
		}
	}
	return bar, found
}

func (ui *taskExecTUI) findStatusBar(ts *taskStatus) (bar int, found bool) {
	for i := 0; i < ui.numStatusBars; i++ {
		if status, ok := ui.statusBars[i]; ok {
			if ts == status {
				bar = i
				found = true
				return bar, found
			}
		}
	}

	return bar, found
}

func (ui *taskExecTUI) TaskStarted(task *executor.Task) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ts, ok := ui.statuses[task]
	if !ok {
		ui.out.Verbose("warning: task not found in internal 'statuses'")
		return
	}

	ts.startedAt = ui.clock()

	// Find free slot
	bar, found := ui.useFreeStatusBar(ts)
	if !found {
		ui.out.Verbose("warning: no free status bar found to display task status")
		return
	}

	ui.progress.StatusBarResetf(bar, ts.displayName, ts.String())
}

func (ui *taskExecTUI) TaskCurrentlyExecuting(task *executor.Task, message string) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ts, ok := ui.statuses[task]
	if !ok {
		ui.out.Verbose("warning: task not found in internal 'statuses'")
		return
	}

	ts.currentlyExecuting = message

	bar, found := ui.findStatusBar(ts)
	if !found {
		ui.out.Verbose("warning: no free status bar found to display task status")
		return
	}

	ui.progress.StatusBarUpdatef(bar, ts.String())
}

func (ui *taskExecTUI) StepsExecutionUI(task *executor.Task) executor.StepsExecutionUI {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ts, ok := ui.statuses[task]
	if !ok {
		ui.out.Verbose("warning: task not found in internal 'statuses'")
		return executor.NoopStepsExecUI{}
	}

	bar, found := ui.findStatusBar(ts)
	if !found {
		ui.out.Verbose("warning: no free status bar found to display task status")
		return executor.NoopStepsExecUI{}
	}

	return &stepsExecTUI{
		task: task,
		updateStatusBar: func(message string) {
			ts.currentlyExecuting = message
			ui.progress.StatusBarUpdatef(bar, ts.String())
		},
	}
}

func (ui *taskExecTUI) TaskFinished(task *executor.Task, err error) {
	ui.mu.Lock()
	defer ui.mu.Unlock()

	ts, ok := ui.statuses[task]
	if !ok {
		ui.out.Verbose("warning: task not found in internal 'statuses'")
		return
	}

	ts.finishedAt = ui.clock()
	ts.err = err

	ui.finished += 1
	if ts.err != nil {
		ui.errored += 1
	}
	ui.updateProgressBar(ui.finished, ui.errored, len(ui.statuses))

	bar, found := ui.findStatusBar(ts)
	if !found {
		ui.out.Verbose("warning: no free status bar found to display task status")
		return
	}

	if ts.err != nil {
		ui.progress.StatusBarFailf(bar, ts.String())
	} else {
		ui.progress.StatusBarCompletef(bar, ts.String())
	}

	delete(ui.statusBars, bar)
}

func (ui *taskExecTUI) TaskChangesetSpecsBuilt(task *executor.Task, specs []*batcheslib.ChangesetSpec) {
	if !ui.verbose {
		return
	}

	ui.mu.Lock()
	defer ui.mu.Unlock()

	ts, ok := ui.statuses[task]
	if !ok {
		ui.out.Verbose("warning: task not found in internal 'statuses'")
		return
	}

	var fileDiffs []*diff.FileDiff
	for _, spec := range specs {
		fd, err := diff.ParseMultiFileDiff(spec.Commits[0].Diff)
		if err != nil {
			ui.progress.Verbosef("%-*s failed to display status: %s", ui.maxRepoName, ts.displayName, err)
			return
		}
		fileDiffs = append(fileDiffs, fd...)
	}

	ui.progress.VerboseLine(output.Linef("", output.StylePending, "%s", ts.displayName))

	if len(fileDiffs) == 0 {
		ui.progress.Verbosef("  No changes")
	} else {
		lines, err := verboseDiffSummary(fileDiffs)
		if err != nil {
			ui.progress.Verbosef("%-*s failed to display status: %s", ui.maxRepoName, ts.displayName, err)
			return
		}

		for _, line := range lines {
			ui.progress.Verbose(line)
		}
	}

	if len(specs) > 1 {
		ui.progress.Verbosef("  %d changeset specs generated", len(specs))
	}
	ui.progress.Verbosef("  Execution took %s", ts.ExecutionTime())
	ui.progress.Verbose("")
}

func (ui *taskExecTUI) updateProgressBar(completed, errored, total int) {
	ui.progress.SetValue(0, float64(completed))

	label := fmt.Sprintf("Executing... (%d/%d, %d errored)", completed, total, errored)
	ui.progress.SetLabelAndRecalc(0, label)
}

type statusTexter interface {
	StatusText() string
}

func verboseDiffSummary(fileDiffs []*diff.FileDiff) ([]string, error) {
	var (
		lines []string

		maxFilenameLen int
		sumInsertions  int
		sumDeletions   int
	)

	fileStats := make(map[string]string, len(fileDiffs))
	fileNames := make([]string, len(fileDiffs))

	for i, f := range fileDiffs {
		name := diffDisplayName(f)

		fileNames[i] = name

		if len(name) > maxFilenameLen {
			maxFilenameLen = len(name)
		}

		stat := f.Stat()

		sumInsertions += int(stat.Added) + int(stat.Changed)
		sumDeletions += int(stat.Deleted) + int(stat.Changed)

		num := stat.Added + 2*stat.Changed + stat.Deleted

		fileStats[name] = fmt.Sprintf("%d %s", num, diffStatDiagram(stat))
	}

	sort.Slice(fileNames, func(i, j int) bool { return fileNames[i] < fileNames[j] })

	for _, name := range fileNames {
		stats := fileStats[name]
		lines = append(lines, fmt.Sprintf("\t%-*s | %s", maxFilenameLen, name, stats))
	}

	var insertionsPlural string
	if sumInsertions != 0 {
		insertionsPlural = "s"
	}

	var deletionsPlural string
	if sumDeletions != 1 {
		deletionsPlural = "s"
	}

	lines = append(lines, fmt.Sprintf("  %s, %s, %s",
		diffStatDescription(fileDiffs),
		fmt.Sprintf("%d insertion%s", sumInsertions, insertionsPlural),
		fmt.Sprintf("%d deletion%s", sumDeletions, deletionsPlural),
	))

	return lines, nil
}

func diffDisplayName(f *diff.FileDiff) string {
	name := f.NewName
	if name == "/dev/null" {
		name = f.OrigName
	}
	return name
}

func diffStatDescription(fileDiffs []*diff.FileDiff) string {
	var plural string
	if len(fileDiffs) > 1 {
		plural = "s"
	}

	return fmt.Sprintf("%d file%s changed", len(fileDiffs), plural)
}

func diffStatDiagram(stat diff.Stat) string {
	const maxWidth = 20
	added := float64(stat.Added + stat.Changed)
	deleted := float64(stat.Deleted + stat.Changed)
	if total := added + deleted; total > maxWidth {
		x := float64(20) / total
		added *= x
		deleted *= x
	}
	return fmt.Sprintf("%s%s%s%s%s",
		output.StyleLinesAdded, strings.Repeat("+", int(added)),
		output.StyleLinesDeleted, strings.Repeat("-", int(deleted)),
		output.StyleReset,
	)
}

type stepsExecTUI struct {
	task            *executor.Task
	updateStatusBar func(string)
}

func (ui stepsExecTUI) ArchiveDownloadStarted() {
	ui.updateStatusBar("Downloading archive")
}
func (ui stepsExecTUI) ArchiveDownloadFinished(err error) {}
func (ui stepsExecTUI) WorkspaceInitializationStarted() {
	ui.updateStatusBar("Initializing workspace")
}
func (ui stepsExecTUI) WorkspaceInitializationFinished() {}
func (ui stepsExecTUI) SkippingStepsUpto(startStep int) {
	switch startStep {
	case 1:
		ui.updateStatusBar("Skipping step 1. Found cached result.")
	default:
		ui.updateStatusBar(fmt.Sprintf("Skipping steps 1 to %d. Found cached results.", startStep))
	}
}

func (ui stepsExecTUI) StepSkipped(step int) {
	ui.updateStatusBar(fmt.Sprintf("Skipping step %d", step))
}
func (ui stepsExecTUI) StepPreparingStart(step int) {
	ui.updateStatusBar(fmt.Sprintf("Preparing step %d", step))
}
func (ui stepsExecTUI) StepPreparingSuccess(step int) {
	// noop right now
}
func (ui stepsExecTUI) StepPreparingFailed(step int, err error) {
	// noop right now
}
func (ui stepsExecTUI) StepStarted(step int, runScript string, _ map[string]string) {
	ui.updateStatusBar(runScript)
}

func (ui stepsExecTUI) StepOutputWriter(ctx context.Context, task *executor.Task, step int) executor.StepOutputWriter {
	return executor.NoopStepOutputWriter{}
}

func (ui stepsExecTUI) StepFinished(idx int, diff []byte, changes git.Changes, outputs map[string]interface{}) {
	// noop right now
}
func (ui stepsExecTUI) StepFailed(idx int, err error, exitCode int) {
	// noop right now
}
