package main

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/fatih/color"
	"github.com/gosuri/uilive"
	"github.com/sourcegraph/go-diff/diff"
)

func newTerminalUI(keepLogs bool) func(reposMap map[ActionRepo]ActionRepoStatus) {
	uilive.Out = os.Stderr
	uilive.RefreshInterval = 10 * time.Hour // TODO!(sqs): manually flush
	color.NoColor = false                   // force color even when in a pipe
	var (
		lwMu sync.Mutex
		lw   = uilive.New()
	)
	lw.Start()
	defer lw.Stop()

	spinner := []rune{'⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'}
	spinnerI := 0
	return func(reposMap map[ActionRepo]ActionRepoStatus) {
		lwMu.Lock()
		defer lwMu.Unlock()

		spinnerRune := spinner[spinnerI%len(spinner)]
		spinnerI++

		reposSorted := make([]ActionRepo, 0, len(reposMap))
		repoNameLen := 0
		for repo := range reposMap {
			reposSorted = append(reposSorted, repo)
			if n := utf8.RuneCountInString(repo.Name); n > repoNameLen {
				repoNameLen = n
			}
		}
		sort.Slice(reposSorted, func(i, j int) bool { return reposSorted[i].Name < reposSorted[j].Name })

		for i, repo := range reposSorted {
			status := reposMap[repo]

			var (
				timerDuration time.Duration

				statusColor func(string, ...interface{}) string

				statusText  string
				logFileText string
			)
			if keepLogs && status.LogFile != "" {
				logFileText = color.HiBlackString(status.LogFile)
			}
			switch {
			case !status.Cached && status.StartedAt.IsZero():
				statusColor = color.HiBlackString
				statusText = statusColor(string(spinnerRune))
				timerDuration = time.Since(status.EnqueuedAt)

			case !status.Cached && status.FinishedAt.IsZero():
				statusColor = color.YellowString
				statusText = statusColor(string(spinnerRune))
				timerDuration = time.Since(status.StartedAt)

			case status.Cached || !status.FinishedAt.IsZero():
				if status.Err != nil {
					statusColor = color.RedString
					statusText = "error: see " + status.LogFile
					logFileText = "" // don't show twice
				} else {
					statusColor = color.GreenString
					if status.Patch != (CampaignPlanPatch{}) && status.Patch.Patch != "" {
						fileDiffs, err := diff.ParseMultiFileDiff([]byte(status.Patch.Patch))
						if err != nil {
							panic(err)
							// return errors.Wrapf(err, "invalid patch for repository %q", repo.Name)
						}
						statusText = diffStatDescription(fileDiffs) + " " + diffStatDiagram(sumDiffStats(fileDiffs))
						if status.Cached {
							statusText += " (cached)"
						}
					} else {
						statusText = color.HiBlackString("0 files changed")
					}
				}
				timerDuration = status.FinishedAt.Sub(status.StartedAt)
			}

			var w io.Writer
			if i == 0 {
				w = lw
			} else {
				w = lw.Newline()
			}

			var appendTexts []string
			if statusText != "" {
				appendTexts = append(appendTexts, statusText)
			}
			if logFileText != "" {
				appendTexts = append(appendTexts, logFileText)
			}
			repoText := statusColor(fmt.Sprintf("%-*s", repoNameLen, repo.Name))
			pipe := color.HiBlackString("|")
			fmt.Fprintf(w, "%s %s ", repoText, pipe)
			fmt.Fprintf(w, "%s", strings.Join(appendTexts, " "))
			if timerDuration != 0 {
				fmt.Fprintf(w, color.HiBlackString(" %s"), timerDuration.Round(time.Second))
			}
			fmt.Fprintln(w)
		}
		_ = lw.Flush()
	}
}
