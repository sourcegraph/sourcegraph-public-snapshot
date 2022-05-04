package repo

import (
	"strings"
	"sync"

	"github.com/sourcegraph/go-diff/diff"

	"github.com/sourcegraph/sourcegraph/dev/sg/internal/run"
)

// State represents the state of the repository.
type State struct {
	// Branch is the currently checked out branch.
	Branch string

	diff     map[string][]DiffHunk
	diffErr  error
	diffOnce sync.Once
}

type DiffHunk struct {
	// StartLine is new start line
	StartLine int
	// AddedLines are lines that got added
	AddedLines []string
}

func (s *State) GetDiff(files string) (map[string][]DiffHunk, error) {
	s.diffOnce.Do(func() {
		if files == "" {
			files = "**/*"
		}

		target := "origin/main"
		if s.Branch == "main" {
			target = "@^"
		}

		diffOutput, err := run.TrimResult(run.GitCmd("diff", target, "--", files))
		if err != nil {
			s.diffErr = err
			return
		}
		s.diff, s.diffErr = parseDiff(diffOutput)
	})
	return s.diff, s.diffErr
}

func parseDiff(diffOutput string) (map[string][]DiffHunk, error) {
	fullDiffs, err := diff.ParseMultiFileDiff([]byte(diffOutput))
	if err != nil {
		return nil, err
	}

	diffs := make(map[string][]DiffHunk)
	for _, d := range fullDiffs {
		if d.NewName == "" {
			continue
		}

		// b/dev/sg/lints.go -> dev/sg/lints.go
		fileName := strings.SplitN(d.NewName, "/", 2)[1]

		// Summarize hunks
		for _, h := range d.Hunks {
			lines := strings.Split(string(h.Body), "\n")

			var addedLines []string
			for _, l := range lines {
				// +$LINE -> $LINE
				if strings.HasPrefix(l, "+") {
					addedLines = append(addedLines, strings.TrimPrefix(l, "+"))
				}
			}

			diffs[fileName] = append(diffs[fileName], DiffHunk{
				StartLine:  int(h.NewStartLine),
				AddedLines: addedLines,
			})
		}
	}
	return diffs, nil
}
