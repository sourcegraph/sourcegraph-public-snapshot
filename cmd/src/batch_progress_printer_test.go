package main

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/src-cli/internal/batches"
	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/output"
)

const progressPrinterDiff = `diff --git README.md README.md
new file mode 100644
index 0000000..3363c39
--- /dev/null
+++ README.md
@@ -0,0 +1,3 @@
+# README
+
+This is the readme
diff --git a/b/c/c.txt a/b/c/c.txt
deleted file mode 100644
index 5da75cf..0000000
--- a/b/c/c.txt
+++ /dev/null
@@ -1 +0,0 @@
-this is c
diff --git x/x.txt x/x.txt
index 627c2ae..88f1836 100644
--- x/x.txt
+++ x/x.txt
@@ -1 +1 @@
-this is x
+this is x (or is it?)
`

func TestBatchProgressPrinterIntegration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Something emits different escape codes on windows.")
	}

	buf := &ttyBuf{}

	out := output.NewOutput(buf, output.OutputOpts{
		ForceTTY:    true,
		ForceColor:  true,
		ForceHeight: 25,
		ForceWidth:  80,
		Verbose:     true,
	})

	now := time.Now()
	statuses := []*executor.TaskStatus{
		{
			RepoName:           "github.com/sourcegraph/sourcegraph",
			StartedAt:          now,
			CurrentlyExecuting: "echo Hello World > README.md",
		},
		{
			RepoName:           "github.com/sourcegraph/src-cli",
			StartedAt:          now.Add(time.Duration(-5) * time.Second),
			CurrentlyExecuting: "Downloading archive",
		},
		{
			RepoName:           "github.com/sourcegraph/automation-testing",
			StartedAt:          now.Add(time.Duration(-5) * time.Second),
			CurrentlyExecuting: "echo Hello World > README.md",
		},
	}

	printer := newBatchProgressPrinter(out, true, 4)
	printer.forceNoSpinner = true

	// Print with all three tasks running
	printer.PrintStatuses(statuses)
	have := buf.Lines()
	want := []string{
		"⠋  Executing... (0/3, 0 errored)  ",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  echo Hello World > README.md          0s",
		"├── github.com/sourcegraph/src-cli      Downloading archive                   0s",
		"└── github.com/sourcegraph/automati...  echo Hello World > README.md          0s",
		"",
	}
	if !cmp.Equal(want, have) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, have))
	}

	// Now mark the last task as completed
	statuses[len(statuses)-1] = &executor.TaskStatus{
		RepoName:           "github.com/sourcegraph/automation-testing",
		StartedAt:          now.Add(time.Duration(-5) * time.Second),
		FinishedAt:         now.Add(time.Duration(5) * time.Second),
		CurrentlyExecuting: "",
		Err:                nil,
		ChangesetSpecs: []*batches.ChangesetSpec{
			{
				BaseRepository: "graphql-id",
				CreatedChangeset: &batches.CreatedChangeset{
					BaseRef:        "refs/heads/main",
					BaseRev:        "d34db33f",
					HeadRepository: "graphql-id",
					HeadRef:        "refs/heads/my-batch-change",
					Title:          "This is my batch change",
					Body:           "This is my batch change",
					Commits: []batches.GitCommitDescription{
						{
							Message: "This is my batch change",
							Diff:    progressPrinterDiff,
						},
					},
					Published: false,
				},
			},
		},
	}

	printer.PrintStatuses(statuses)
	have = buf.Lines()
	want = []string{
		"github.com/sourcegraph/automation-testing",
		"\tREADME.md   | 3 +++",
		"\ta/b/c/c.txt | 1 -",
		"\tx/x.txt     | 2 +-",
		"  3 files changed, 4 insertions, 2 deletions",
		"  Execution took 10s",
		"",
		"⠋  Executing... (1/3, 0 errored)  ███████████████▍",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  echo Hello World > README.md          0s",
		"├── github.com/sourcegraph/src-cli      Downloading archive                   0s",
		"└── github.com/sourcegraph/automati...  3 files changed ++++               0s",
		"",
	}
	if !cmp.Equal(want, have) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, have))
	}

	// Print again to make sure we get the same result
	printer.PrintStatuses(statuses)
	have = buf.Lines()
	if !cmp.Equal(want, have) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, have))
	}
}

type ttyBuf struct {
	lines [][]byte

	line   int
	column int
}

func (t *ttyBuf) Write(b []byte) (int, error) {
	var cur int

	for cur < len(b) {
		switch b[cur] {
		case '\n':
			t.line++
			t.column = 0

			if len(t.lines) == t.line {
				t.lines = append(t.lines, []byte{})
			}

		case '\x1b':
			// Check if we're looking at a VT100 escape code.
			if len(b) <= cur || b[cur+1] != '[' {
				t.writeToCurrentLine(b[cur])
				cur++
				continue
			}

			// First of all: forgive me.
			//
			// Now. Looks like we ran into a VT100 escape code.
			// They follow this structure:
			//
			//      \x1b [ <digit> <command>
			//
			// So we jump over the \x1b[ and try to parse the digit.

			cur = cur + 2 // cur == '\x1b', cur + 1 == '['

			digitStart := cur
			for isDigit(b[cur]) {
				cur++
			}

			rawDigit := string(b[digitStart:cur])
			digit, err := strconv.ParseInt(rawDigit, 0, 64)
			if err != nil {
				return 0, err
			}

			command := b[cur]

			// Debug helper:
			// fmt.Printf("command=%q, digit=%d (t.line=%d, t.column=%d)\n", command, digit, t.line, t.column)

			switch command {
			case 'K':
				// reset current line
				if len(t.lines) > t.line {
					t.lines[t.line] = []byte{}
					t.column = 0
				}
			case 'A':
				// move line up by <digit>
				t.line = t.line - int(digit)

			case 'D':
				// *d*elete cursor by <digit> amount
				t.column = t.column - int(digit)
				if t.column < 0 {
					t.column = 0
				}

			case 'm':
				// noop

			case ';':
				// color, skip over until end of color command
				for b[cur] != 'm' {
					cur++
				}
			}

		default:
			t.writeToCurrentLine(b[cur])
		}

		cur++
	}

	return len(b), nil
}

func (t *ttyBuf) writeToCurrentLine(b byte) {
	if len(t.lines) == t.line {
		t.lines = append(t.lines, []byte{})
	}

	if len(t.lines[t.line]) <= t.column {
		t.lines[t.line] = append(t.lines[t.line], b)
	} else {
		t.lines[t.line][t.column] = b
	}
	t.column++
}

func (t *ttyBuf) Lines() []string {
	var lines []string
	for _, l := range t.lines {
		lines = append(lines, string(l))
	}
	return lines
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}
