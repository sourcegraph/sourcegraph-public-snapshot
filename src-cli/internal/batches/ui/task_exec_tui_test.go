package ui

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/output"

	"github.com/sourcegraph/src-cli/internal/batches/executor"
	"github.com/sourcegraph/src-cli/internal/batches/graphql"
)

var progressPrinterDiff = []byte(`diff --git README.md README.md
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
`)

func TestTaskExecTUI_Integration(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Something emits different escape codes on windows.")
	}

	now := time.Now()
	clock := func() time.Time { return now.UTC().Truncate(time.Millisecond) }
	advanceClock := func(d time.Duration) { now = now.Add(d) }

	buf := &ttyBuf{}

	out := output.NewOutput(buf, output.OutputOpts{
		ForceTTY:    true,
		ForceColor:  true,
		ForceHeight: 25,
		ForceWidth:  80,
		Verbose:     true,
	})

	tasks := []*executor.Task{
		{Repository: &graphql.Repository{Name: "github.com/sourcegraph/sourcegraph"}},
		{Repository: &graphql.Repository{Name: "github.com/sourcegraph/src-cli"}},
		{Repository: &graphql.Repository{Name: "github.com/sourcegraph/automation-testing"}},
		{Repository: &graphql.Repository{Name: "github.com/sourcegraph/tiny-go-repo"}},
	}

	printer := newTaskExecTUI(out, true, 3)
	printer.forceNoSpinner = true
	printer.clock = clock

	// Setup internal state
	printer.Start(tasks)

	// Start the first 3 tasks
	printer.TaskStarted(tasks[0])
	printer.TaskStarted(tasks[1])
	printer.TaskStarted(tasks[2])

	expectOutput(t, buf, []string{
		"⠋  Executing... (0/4, 0 errored)                                              0%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  ...                                   0s",
		"├── github.com/sourcegraph/src-cli      ...                                   0s",
		"└── github.com/sourcegraph/automati...  ...                                   0s",
		"",
	})

	// Update the currently executing of the first 3
	printer.TaskCurrentlyExecuting(tasks[0], "echo Hello World > README.md")
	printer.TaskCurrentlyExecuting(tasks[1], "Downloading archive")
	printer.TaskCurrentlyExecuting(tasks[2], "echo Hello World > README.md")

	expectOutput(t, buf, []string{
		"⠋  Executing... (0/4, 0 errored)                                              0%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  echo Hello World > README.md          0s",
		"├── github.com/sourcegraph/src-cli      Downloading archive                   0s",
		"└── github.com/sourcegraph/automati...  echo Hello World > README.md          0s",
		"",
	})

	// Update it again
	printer.TaskCurrentlyExecuting(tasks[0], "gofmt")
	printer.TaskCurrentlyExecuting(tasks[1], "echo Hello World > README.md")
	printer.TaskCurrentlyExecuting(tasks[2], "echo Hello World > README.md")

	expectOutput(t, buf, []string{
		"⠋  Executing... (0/4, 0 errored)                                              0%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  gofmt                                 0s",
		"├── github.com/sourcegraph/src-cli      echo Hello World > README.md          0s",
		"└── github.com/sourcegraph/automati...  echo Hello World > README.md          0s",
		"",
	})

	// Now mark the last task as finished-execution
	advanceClock(10 * time.Second)
	printer.TaskFinished(tasks[2], nil)

	expectOutput(t, buf, []string{
		"⠋  Executing... (1/4, 0 errored)  ██████████▎                                25%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  gofmt                                 0s",
		"├── github.com/sourcegraph/src-cli      echo Hello World > README.md          0s",
		"└── github.com/sourcegraph/automati...  Done!                                 0s",
		"",
	})

	// Now mark the first task as finished-execution
	advanceClock(5 * time.Second)
	printer.TaskFinished(tasks[0], nil)

	expectOutput(t, buf, []string{
		"⠋  Executing... (2/4, 0 errored)  ████████████████████▌                      50%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  Done!                                 0s",
		"├── github.com/sourcegraph/src-cli      echo Hello World > README.md          0s",
		"└── github.com/sourcegraph/automati...  Done!                                 0s",
		"",
	})

	// Mark the last task as finished-building-specs
	printer.TaskChangesetSpecsBuilt(tasks[2], []*batcheslib.ChangesetSpec{
		{
			BaseRepository: "graphql-id",

			BaseRef:        "refs/heads/main",
			BaseRev:        "d34db33f",
			HeadRepository: "graphql-id",
			HeadRef:        "refs/heads/my-batch-change",
			Title:          "This is my batch change",
			Body:           "This is my batch change",
			Commits: []batcheslib.GitCommitDescription{
				{
					Version: 2,
					Message: "This is my batch change",
					Diff:    progressPrinterDiff,
				},
			},
			Published: batcheslib.PublishedValue{Val: false},
		},
	})

	expectOutput(t, buf, []string{
		"github.com/sourcegraph/automation-testing",
		"\tREADME.md   | 3 +++",
		"\ta/b/c/c.txt | 1 -",
		"\tx/x.txt     | 2 +-",
		"  3 files changed, 4 insertions, 2 deletions",
		"  Execution took 10s",
		"",
		"⠋  Executing... (2/4, 0 errored)  ████████████████████▌                      50%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  Done!                                 0s",
		"├── github.com/sourcegraph/src-cli      echo Hello World > README.md          0s",
		"└── github.com/sourcegraph/automati...  Done!                                 0s",
		"",
	})

	// Now we start the 4th task
	printer.TaskStarted(tasks[3])
	printer.TaskCurrentlyExecuting(tasks[3], "rm -rf ~/.horse-ascii-art")

	expectOutput(t, buf, []string{
		"github.com/sourcegraph/automation-testing",
		"\tREADME.md   | 3 +++",
		"\ta/b/c/c.txt | 1 -",
		"\tx/x.txt     | 2 +-",
		"  3 files changed, 4 insertions, 2 deletions",
		"  Execution took 10s",
		"",
		"⠋  Executing... (2/4, 0 errored)  ████████████████████▌                      50%",
		"│                                                                               ",
		"├── github.com/sourcegraph/tiny-go-...  rm -rf ~/.horse-ascii-art             0s",
		"├── github.com/sourcegraph/src-cli      echo Hello World > README.md          0s",
		"└── github.com/sourcegraph/automati...  Done!                                 0s",
		"",
	})
}

func TestProgressUpdateAfterComplete(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Something emits different escape codes on windows.")
	}

	buf := &ttyBuf{}

	now := time.Now()
	clock := func() time.Time { return now.UTC().Truncate(time.Millisecond) }

	out := output.NewOutput(buf, output.OutputOpts{
		ForceTTY:    true,
		ForceColor:  true,
		ForceHeight: 25,
		ForceWidth:  80,
		Verbose:     true,
	})

	tasks := []*executor.Task{
		{Repository: &graphql.Repository{Name: "github.com/sourcegraph/sourcegraph"}},
		{Repository: &graphql.Repository{Name: "github.com/sourcegraph/src-cli"}},
	}

	printer := newTaskExecTUI(out, true, 2)
	printer.forceNoSpinner = true
	printer.clock = clock

	// Setup internal state.
	printer.Start(tasks)

	// Start the tasks.
	printer.TaskStarted(tasks[0])
	printer.TaskStarted(tasks[1])

	// Update the tasks into a useful state.
	printer.TaskCurrentlyExecuting(tasks[0], "echo Hello World > README.md")
	printer.TaskCurrentlyExecuting(tasks[1], "Downloading archive")

	expectOutput(t, buf, []string{
		"⠋  Executing... (0/2, 0 errored)                                              0%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  echo Hello World > README.md          0s",
		"└── github.com/sourcegraph/src-cli      Downloading archive                   0s",
		"",
	})

	// Now mark the progress as complete.
	printer.progress.Complete()

	// Now send another update. This would panic before the relevant fix was
	// merged in #666.
	printer.TaskCurrentlyExecuting(tasks[0], "exit 42")

	// The actual output is slightly less important at this point, but let's
	// check it anyway.
	expectOutput(t, buf, []string{
		"✅ Executing... (0/2, 0 errored)  ████████████████████████████████████████  100%",
		"│                                                                               ",
		"├── github.com/sourcegraph/sourcegraph  exit 42                               0s",
		"└── github.com/sourcegraph/src-cli      Downloading archive                   0s",
		"",
	})
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

func expectOutput(t *testing.T, buf *ttyBuf, want []string) {
	t.Helper()

	have := buf.Lines()
	if !cmp.Equal(want, have) {
		t.Fatalf("wrong output:\n%s", cmp.Diff(want, have))
	}
}
