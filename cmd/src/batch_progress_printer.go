package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/output"
	"golang.org/x/sync/semaphore"
)

func newBatchProgressPrinter(out *output.Output, verbose bool, numParallelism int) *batchProgressPrinter {
	return &batchProgressPrinter{
		out: out,

		sem: semaphore.NewWeighted(1),

		verbose: verbose,

		numParallelism: numParallelism,

		completedTasks: map[string]bool{},
		runningTasks:   map[string]*executor.TaskStatus{},

		repoStatusBar: map[string]int{},
		statusBarRepo: map[int]string{},
	}
}

type batchProgressPrinter struct {
	// Used in tests only
	forceNoSpinner bool

	out *output.Output

	sem *semaphore.Weighted

	verbose bool

	progress      output.ProgressWithStatusBars
	numStatusBars int

	maxRepoName    int
	numParallelism int

	completedTasks map[string]bool
	runningTasks   map[string]*executor.TaskStatus

	repoStatusBar map[string]int
	statusBarRepo map[int]string
}

func (p *batchProgressPrinter) initProgressBar(statuses []*executor.TaskStatus) int {
	numStatusBars := p.numParallelism
	if len(statuses) < numStatusBars {
		numStatusBars = len(statuses)
	}

	statusBars := make([]*output.StatusBar, 0, numStatusBars)
	for i := 0; i < numStatusBars; i++ {
		statusBars = append(statusBars, output.NewStatusBar())
	}

	progressBars := []output.ProgressBar{
		{
			Label: fmt.Sprintf("Executing ... (0/%d, 0 errored)", len(statuses)),
			Max:   float64(len(statuses)),
		},
	}

	opts := output.DefaultProgressTTYOpts.WithNoSpinner(p.forceNoSpinner)
	p.progress = p.out.ProgressWithStatusBars(progressBars, statusBars, opts)

	return numStatusBars
}

func (p *batchProgressPrinter) Complete() {
	if p.progress != nil {
		p.progress.Complete()
	}
}

func (p *batchProgressPrinter) updateProgressBar(completed, errored, total int) {
	if p.progress == nil {
		return
	}

	p.progress.SetValue(0, float64(completed))

	label := fmt.Sprintf("Executing... (%d/%d, %d errored)", completed, total, errored)
	p.progress.SetLabelAndRecalc(0, label)
}

func (p *batchProgressPrinter) PrintStatuses(statuses []*executor.TaskStatus) {
	if len(statuses) == 0 {
		return
	}

	// Try to acquire semaphore. If that fails, another PrintStatuses is still
	// running and we return, since it will be called again.
	if !p.sem.TryAcquire(1) {
		return
	}
	defer p.sem.Release(1)

	if p.progress == nil {
		p.numStatusBars = p.initProgressBar(statuses)
	}

	newlyCompleted := []*executor.TaskStatus{}
	currentlyRunning := []*executor.TaskStatus{}
	errored := 0

	for _, ts := range statuses {
		if len(ts.DisplayName()) > p.maxRepoName {
			p.maxRepoName = len(ts.DisplayName())
		}

		if ts.IsCompleted() {
			if ts.Err != nil {
				errored += 1
			}

			if !p.completedTasks[ts.DisplayName()] {
				p.completedTasks[ts.DisplayName()] = true
				newlyCompleted = append(newlyCompleted, ts)
			}

			if _, ok := p.runningTasks[ts.DisplayName()]; ok {
				delete(p.runningTasks, ts.DisplayName())

				// Free slot
				idx := p.repoStatusBar[ts.DisplayName()]
				delete(p.statusBarRepo, idx)
			}
		}

		if ts.IsRunning() {
			currentlyRunning = append(currentlyRunning, ts)
		}

	}

	p.updateProgressBar(len(p.completedTasks), errored, len(statuses))

	newlyStarted := map[string]*executor.TaskStatus{}
	statusBarIndex := 0
	for _, ts := range currentlyRunning {
		if _, ok := p.runningTasks[ts.DisplayName()]; ok {
			// Update the status
			p.runningTasks[ts.DisplayName()] = ts
			continue
		}

		p.runningTasks[ts.DisplayName()] = ts
		newlyStarted[ts.DisplayName()] = ts

		// Find free status bar slot
		_, ok := p.statusBarRepo[statusBarIndex]
		for ok {
			statusBarIndex += 1
			_, ok = p.statusBarRepo[statusBarIndex]
		}

		if statusBarIndex >= p.numStatusBars {
			// If the only free slot is past the number of status bars we
			// have, there's a race condition going on where we have more tasks
			// reporting as "currently executing" than could be executing, most
			// likely because one of them hasn't been updated yet.
			break
		}

		p.statusBarRepo[statusBarIndex] = ts.DisplayName()
		p.repoStatusBar[ts.DisplayName()] = statusBarIndex
	}

	for _, ts := range newlyCompleted {
		fileDiffs, hasDiffs, err := ts.FileDiffs()
		if err != nil {
			p.progress.Verbosef("%-*s failed to display status: %s", p.maxRepoName, ts.DisplayName(), err)
			continue
		}

		if p.verbose {
			p.progress.WriteLine(output.Linef("", output.StylePending, "%s", ts.DisplayName()))

			if !hasDiffs {
				p.progress.Verbosef("  No changes")
			} else {
				lines, err := verboseDiffSummary(fileDiffs)
				if err != nil {
					p.progress.Verbosef("%-*s failed to display status: %s", p.maxRepoName, ts.DisplayName(), err)
					continue
				}

				for _, line := range lines {
					p.progress.Verbose(line)
				}
			}

			if len(ts.ChangesetSpecs) > 1 {
				p.progress.Verbosef("  %d changeset specs generated", len(ts.ChangesetSpecs))
			}
			p.progress.Verbosef("  Execution took %s", ts.ExecutionTime())
			p.progress.Verbose("")
		}

		if idx, ok := p.repoStatusBar[ts.DisplayName()]; ok {
			// Log that this task completed, but only if there is no
			// currently executing one in this bar, to avoid flicker.
			if _, ok := p.statusBarRepo[idx]; !ok {
				statusText, err := taskStatusBarText(ts)
				if err != nil {
					p.progress.Verbosef("%-*s failed to display status: %s", p.maxRepoName, ts.DisplayName(), err)
					continue
				}

				if ts.Err != nil {
					p.progress.StatusBarFailf(idx, statusText)
				} else {
					p.progress.StatusBarCompletef(idx, statusText)
				}
			}
			delete(p.repoStatusBar, ts.DisplayName())
		}
	}

	for statusBar, repo := range p.statusBarRepo {
		ts, ok := p.runningTasks[repo]
		if !ok {
			// This should not happen
			continue
		}

		statusText, err := taskStatusBarText(ts)
		if err != nil {
			p.progress.Verbosef("%-*s failed to display status: %s", p.maxRepoName, ts.DisplayName(), err)
			continue
		}

		if _, ok := newlyStarted[repo]; ok {
			p.progress.StatusBarResetf(statusBar, ts.DisplayName(), statusText)
		} else {
			p.progress.StatusBarUpdatef(statusBar, statusText)
		}
	}
}

type statusTexter interface {
	StatusText() string
}

func taskStatusBarText(ts *executor.TaskStatus) (string, error) {
	var statusText string

	if ts.IsCompleted() {
		diffs, hasDiffs, err := ts.FileDiffs()
		if err != nil {
			return "", err
		}

		if hasDiffs {
			statusText = diffStatDescription(diffs) + " " + diffStatDiagram(sumDiffStats(diffs))
		} else {
			if ts.Err != nil {
				if texter, ok := ts.Err.(statusTexter); ok {
					statusText = texter.StatusText()
				} else {
					statusText = ts.Err.Error()
				}
			} else {
				statusText = "No changes"
			}
		}

		if ts.Cached {
			statusText += " (cached)"
		}
	} else if ts.IsRunning() {
		if ts.CurrentlyExecuting != "" {
			lines := strings.Split(ts.CurrentlyExecuting, "\n")
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

	return statusText, nil
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

func sumDiffStats(fileDiffs []*diff.FileDiff) diff.Stat {
	sum := diff.Stat{}
	for _, fileDiff := range fileDiffs {
		stat := fileDiff.Stat()
		sum.Added += stat.Added
		sum.Changed += stat.Changed
		sum.Deleted += stat.Deleted
	}
	return sum
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
