package main

import (
	"fmt"
	"strings"

	"github.com/sourcegraph/go-diff/diff"
	"github.com/sourcegraph/src-cli/internal/campaigns"
	"github.com/sourcegraph/src-cli/internal/output"
)

func newCampaignProgressPrinter(out *output.Output, numParallelism int) *campaignProgressPrinter {
	return &campaignProgressPrinter{
		out: out,

		numParallelism: numParallelism,

		completedTasks: map[string]bool{},
		runningTasks:   map[string]*campaigns.TaskStatus{},

		repoStatusBar: map[string]int{},
		statusBarRepo: map[int]string{},
	}
}

type campaignProgressPrinter struct {
	out      *output.Output
	progress output.ProgressWithStatusBars

	maxRepoName    int
	numParallelism int

	completedTasks map[string]bool
	runningTasks   map[string]*campaigns.TaskStatus

	repoStatusBar map[string]int
	statusBarRepo map[int]string
}

func (p *campaignProgressPrinter) initProgressBar(statuses []*campaigns.TaskStatus) {
	numStatusBars := p.numParallelism
	if len(statuses) < numStatusBars {
		numStatusBars = len(statuses)
	}

	statusBars := make([]*output.StatusBar, 0, numStatusBars)
	for i := 0; i < numStatusBars; i++ {
		statusBars = append(statusBars, output.NewStatusBar())
	}

	p.progress = p.out.ProgressWithStatusBars([]output.ProgressBar{{
		Label: fmt.Sprintf("Executing steps in %d repositories", len(statuses)),
		Max:   float64(len(statuses)),
	}}, statusBars, nil)
}

func (p *campaignProgressPrinter) Complete() {
	if p.progress != nil {
		p.progress.Complete()
	}
}

func (p *campaignProgressPrinter) PrintStatuses(statuses []*campaigns.TaskStatus) {
	if len(statuses) == 0 {
		return
	}

	if p.progress == nil {
		p.initProgressBar(statuses)
	}

	newlyCompleted := []*campaigns.TaskStatus{}
	currentlyRunning := []*campaigns.TaskStatus{}

	for _, ts := range statuses {
		if len(ts.RepoName) > p.maxRepoName {
			p.maxRepoName = len(ts.RepoName)
		}

		if ts.IsCompleted() {
			if !p.completedTasks[ts.RepoName] {
				p.completedTasks[ts.RepoName] = true
				newlyCompleted = append(newlyCompleted, ts)
			}

			if _, ok := p.runningTasks[ts.RepoName]; ok {
				delete(p.runningTasks, ts.RepoName)

				// Free slot
				idx := p.repoStatusBar[ts.RepoName]
				delete(p.statusBarRepo, idx)
			}
		}

		if ts.IsRunning() {
			currentlyRunning = append(currentlyRunning, ts)
		}

	}

	p.progress.SetValue(0, float64(len(p.completedTasks)))

	newlyStarted := map[string]*campaigns.TaskStatus{}
	statusBarIndex := 0
	for _, ts := range currentlyRunning {
		if _, ok := p.runningTasks[ts.RepoName]; ok {
			continue
		}

		newlyStarted[ts.RepoName] = ts
		p.runningTasks[ts.RepoName] = ts

		// Find free slot
		_, ok := p.statusBarRepo[statusBarIndex]
		for ok {
			statusBarIndex += 1
			_, ok = p.statusBarRepo[statusBarIndex]
		}

		p.statusBarRepo[statusBarIndex] = ts.RepoName
		p.repoStatusBar[ts.RepoName] = statusBarIndex
	}

	for _, ts := range newlyCompleted {
		statusText, err := taskStatusText(ts)
		if err != nil {
			p.progress.Verbosef("%-*s failed to display status: %s", p.maxRepoName, ts.RepoName, err)
			continue
		}

		p.progress.Verbosef("%-*s %s", p.maxRepoName, ts.RepoName, statusText)

		if idx, ok := p.repoStatusBar[ts.RepoName]; ok {
			// Log that this task completed, but only if there is no
			// currently executing one in this bar, to avoid flicker.
			if _, ok := p.statusBarRepo[idx]; !ok {
				if ts.Err != nil {
					p.progress.StatusBarFailf(idx, statusText)
				} else {
					p.progress.StatusBarCompletef(idx, statusText)
				}
			}
			delete(p.repoStatusBar, ts.RepoName)
		}
	}

	for statusBar, repo := range p.statusBarRepo {
		ts, ok := p.runningTasks[repo]
		if !ok {
			// This should not happen
			continue
		}

		statusText, err := taskStatusText(ts)
		if err != nil {
			p.progress.Verbosef("%-*s failed to display status: %s", p.maxRepoName, ts.RepoName, err)
			continue
		}

		if _, ok := newlyStarted[repo]; ok {
			p.progress.StatusBarResetf(statusBar, ts.RepoName, statusText)
		} else {
			p.progress.StatusBarUpdatef(statusBar, statusText)
		}
	}
}

type statusTexter interface {
	StatusText() string
}

func taskStatusText(ts *campaigns.TaskStatus) (string, error) {
	var statusText string

	if ts.IsCompleted() {
		if ts.ChangesetSpec == nil {
			if ts.Err != nil {
				if texter, ok := ts.Err.(statusTexter); ok {
					statusText = texter.StatusText()
				} else {
					statusText = ts.Err.Error()
				}
			} else {
				statusText = "No changes"
			}
		} else {
			fileDiffs, err := diff.ParseMultiFileDiff([]byte(ts.ChangesetSpec.Commits[0].Diff))
			if err != nil {
				return "", err
			}

			statusText = diffStatDescription(fileDiffs) + " " + diffStatDiagram(sumDiffStats(fileDiffs))
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
				statusText = fmt.Sprintf("%s", escapedLine)
			}
		} else {
			statusText = fmt.Sprintf("...")
		}
	}

	return statusText, nil
}
