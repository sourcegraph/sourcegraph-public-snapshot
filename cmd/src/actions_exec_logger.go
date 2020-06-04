package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/neelance/parallel"
	"github.com/pkg/errors"
	"github.com/segmentio/textio"
)

var (
	boldBlack = color.New(color.Bold, color.FgBlack)
	boldRed   = color.New(color.Bold, color.FgRed)
	boldGreen = color.New(color.Bold, color.FgGreen)
	hiGreen   = color.New(color.FgHiGreen)
	yellow    = color.New(color.FgYellow)
	grey      = color.New(color.FgHiBlack)
)

type actionLogger struct {
	verbose  bool
	keepLogs bool

	highlight func(a ...interface{}) string

	progress *progress
	out      io.WriteCloser

	mu         sync.Mutex
	logFiles   map[string]*os.File
	logWriters map[string]io.Writer
}

func newActionLogger(verbose, keepLogs bool) *actionLogger {
	useColor := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
	if useColor {
		color.NoColor = false
	}

	progress := new(progress)

	return &actionLogger{
		verbose:   verbose,
		keepLogs:  keepLogs,
		highlight: color.New(color.Bold, color.FgGreen).SprintFunc(),
		progress:  progress,
		out: &progressWriter{
			p: progress,
			w: os.Stderr,
		},
		logFiles:   map[string]*os.File{},
		logWriters: map[string]io.Writer{},
	}
}

func (a *actionLogger) Start(totalSteps int) {
	if a.verbose {
		a.progress.SetTotalSteps(int64(totalSteps))
		fmt.Fprintln(os.Stderr)
	}
}

func (a *actionLogger) Infof(format string, args ...interface{}) {
	a.log("", grey, format, args...)
}

func (a *actionLogger) Warnf(format string, args ...interface{}) {
	a.log("", yellow, "WARNING: "+format, args...)
}

func (a *actionLogger) ActionFailed(err error, patches []PatchInput) {
	if !a.verbose {
		return
	}
	a.out.Close()
	fmt.Fprintln(os.Stderr)
	if perr, ok := err.(parallel.Errors); ok {
		if len(patches) > 0 {
			yellow.Fprintf(os.Stderr, "✗  Action produced %d patches but failed with %d errors:\n\n", len(patches), len(perr))
		} else {
			yellow.Fprintf(os.Stderr, "✗  Action failed with %d errors:\n", len(perr))
		}
		for _, e := range perr {
			fmt.Fprintf(os.Stderr, "\t- %s\n", e)
		}
		fmt.Println()
	} else if err != nil {
		if len(patches) > 0 {
			yellow.Fprintf(os.Stderr, "✗  Action produced %d patches but failed with error: %s\n\n", len(patches), err)
		} else {
			yellow.Fprintf(os.Stderr, "✗  Action failed with error: %s\n\n", err)
		}
	} else {
		grey.Fprintf(os.Stderr, "✗  Action did not produce any patches.\n\n")
	}
}

func (a *actionLogger) ActionSuccess(patches []PatchInput, newLines bool) {
	if !a.verbose {
		return
	}
	a.out.Close()
	fmt.Fprintln(os.Stderr)
	format := "✔  Action produced %d patches."
	if newLines {
		format = format + "\n\n"
	}
	hiGreen.Fprintf(os.Stderr, format, len(patches))
}

func (a *actionLogger) RepoCacheHit(repo ActionRepo, stepCount int, patchProduced bool) {
	a.progress.IncStepsComplete(int64(stepCount))
	if patchProduced {
		a.progress.IncPatchCount()
		a.log(repo.Name, boldGreen, "Cached result found: using cached diff.\n")
		return
	}
	a.log(repo.Name, grey, "Cached result found: no diff produced for this repository.\n")
}

func (a *actionLogger) AddRepo(repo ActionRepo) (string, error) {
	prefix := "action-" + strings.Replace(strings.Replace(repo.Name, "/", "-", -1), "github.com-", "", -1)

	logFile, err := ioutil.TempFile(tempDirPrefix, prefix+"-log")
	if err != nil {
		return "", err
	}

	logWriter := io.Writer(logFile)

	a.mu.Lock()
	defer a.mu.Unlock()
	a.logFiles[repo.Name] = logFile
	a.logWriters[repo.Name] = logWriter

	return logFile.Name(), nil
}

func (a *actionLogger) RepoWriter(repoName string) (io.Writer, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w, ok := a.logWriters[repoName]
	return w, ok
}

func (a *actionLogger) RepoStdoutStderr(repoName string) (io.Writer, io.Writer, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w, ok := a.logWriters[repoName]

	if !a.verbose {
		return w, w, ok
	}

	stderrPrefix := fmt.Sprintf("%s -> [STDERR]: ", yellow.Sprint(repoName))
	stderr := textio.NewPrefixWriter(a.out, stderrPrefix)

	stdoutPrefix := fmt.Sprintf("%s -> [STDOUT]: ", yellow.Sprint(repoName))
	stdout := textio.NewPrefixWriter(a.out, stdoutPrefix)

	return io.MultiWriter(stdout, w), io.MultiWriter(stderr, w), ok
}

func (a *actionLogger) RepoFinished(repoName string, patchProduced bool, actionErr error) error {
	a.mu.Lock()
	f, ok := a.logFiles[repoName]
	if !ok {
		a.mu.Unlock()
		return nil
	}
	a.mu.Unlock()

	if actionErr != nil {
		if a.keepLogs {
			a.write(repoName, boldRed, "Action failed: %q (Logfile: %s)\n", actionErr, f.Name())
		} else {
			a.write(repoName, boldRed, "Action failed: %q\n", actionErr)
		}
	} else if patchProduced {
		a.progress.IncPatchCount()
		a.write(repoName, boldGreen, "Finished. Patch produced.\n")
	} else {
		a.write(repoName, grey, "Finished. No patch produced.\n")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	delete(a.logFiles, repoName)
	delete(a.logWriters, repoName)

	if !a.keepLogs && actionErr == nil {
		if err := os.Remove(f.Name()); err != nil {
			return errors.Wrap(err, "Failed to remove log file")
		}
	}

	return nil
}

func (a *actionLogger) RepoStarted(repoName, rev string, steps []*ActionStep) {
	a.write(repoName, yellow, "Starting action @ %s (%d steps)\n", rev, len(steps))
}

func (a *actionLogger) CommandStepStarted(repoName string, step int, args []string) {
	a.write(repoName, yellow, "%s command %v\n", boldBlack.Sprintf("[Step %d]", step), args)
}

func (a *actionLogger) CommandStepErrored(repoName string, step int, err error) {
	a.progress.IncStepsComplete(1)
	a.progress.IncStepsFailed()
	a.write(repoName, boldRed, "%s %s.\n", boldBlack.Sprintf("[Step %d]", step), err)
}

func (a *actionLogger) CommandStepDone(repoName string, step int) {
	a.progress.IncStepsComplete(1)
	a.write(repoName, yellow, "%s Done.\n", boldBlack.Sprintf("[Step %d]", step))
}

func (a *actionLogger) DockerStepStarted(repoName string, step int, image string) {
	a.write(repoName, yellow, "%s docker run %s\n", boldBlack.Sprintf("[Step %d]", step), image)
}

func (a *actionLogger) DockerStepErrored(repoName string, step int, err error, elapsed time.Duration) {
	a.progress.IncStepsComplete(1)
	a.progress.IncStepsFailed()
	a.write(repoName, boldRed, "%s %s. (%s)\n", boldBlack.Sprintf("[Step %d]", step), err, elapsed)
}

func (a *actionLogger) DockerStepDone(repoName string, step int, elapsed time.Duration) {
	a.progress.IncStepsComplete(1)
	a.write(repoName, yellow, "%s Done. (%s)\n", boldBlack.Sprintf("[Step %d]", step), elapsed)
}

// write writes to the RepoWriter associated with the given repoName and logs the message using the log method.
func (a *actionLogger) write(repoName string, c *color.Color, format string, args ...interface{}) {
	if w, ok := a.RepoWriter(repoName); ok {
		fmt.Fprintf(w, format, args...)
	}
	a.log(repoName, c, format, args...)
}

// log logs only to stderr, it does not log to our repoWriters. When not in verbose mode, it's a noop.
func (a *actionLogger) log(repoName string, c *color.Color, format string, args ...interface{}) {
	if !a.verbose {
		return
	}
	if len(repoName) > 0 {
		format = fmt.Sprintf("%s -> %s", c.Sprint(repoName), format)
	}
	fmt.Fprintf(a.out, format, args...)
}

type progress struct {
	patchCount int64

	totalSteps    int64
	stepsComplete int64
	stepsFailed   int64
}

func (p *progress) SetTotalSteps(n int64) {
	atomic.StoreInt64(&p.totalSteps, n)
}

func (p *progress) TotalSteps() int64 {
	return atomic.LoadInt64(&p.totalSteps)
}

func (p *progress) StepsComplete() int64 {
	return atomic.LoadInt64(&p.stepsComplete)
}

func (p *progress) IncStepsComplete(delta int64) {
	atomic.AddInt64(&p.stepsComplete, delta)
}

func (p *progress) TotalStepsFailed() int64 {
	return atomic.LoadInt64(&p.stepsFailed)
}

func (p *progress) IncStepsFailed() {
	atomic.AddInt64(&p.stepsFailed, 1)
}

func (p *progress) PatchCount() int64 {
	return atomic.LoadInt64(&p.patchCount)
}

func (p *progress) IncPatchCount() {
	atomic.AddInt64(&p.patchCount, 1)
}

type progressWriter struct {
	p *progress

	mu                sync.Mutex
	w                 io.Writer
	shouldClear       bool
	progressLogLength int
	closed            bool
}

func (w *progressWriter) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, fmt.Errorf("writer closed")
	}
	w.clear()

	if w.p.TotalSteps() == 0 {
		// Don't display bar until we know number of steps
		w.shouldClear = false
		return w.w.Write(data)
	}

	if !bytes.HasSuffix(data, []byte("\n")) {
		w.shouldClear = false
		return w.w.Write(data)
	}

	n, err := w.w.Write(data)
	if err != nil {
		return n, err
	}
	total := w.p.TotalSteps()
	done := w.p.StepsComplete()
	var pctDone float64
	if total > 0 {
		pctDone = float64(done) / float64(total)
	}

	maxLength := 50
	bar := strings.Repeat("=", int(float64(maxLength)*pctDone))
	if len(bar) < maxLength {
		bar += ">"
	}
	bar += strings.Repeat(" ", maxLength-len(bar))
	progessText := fmt.Sprintf("[%s] Steps: %d/%d (%s, %s)", bar, w.p.StepsComplete(), w.p.TotalSteps(), boldRed.Sprintf("%d failed", w.p.TotalStepsFailed()), hiGreen.Sprintf("%d patches", w.p.PatchCount()))
	fmt.Fprintf(w.w, progessText)
	w.shouldClear = true
	w.progressLogLength = len(progessText)
	return n, err
}

// Close clears the progress bar and disallows further writing
func (w *progressWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.clear()
	w.closed = true
	return nil
}

func (w *progressWriter) clear() {
	if !w.shouldClear {
		return
	}
	fmt.Fprintf(w.w, "\r")
	fmt.Fprintf(w.w, strings.Repeat(" ", w.progressLogLength))
	fmt.Fprintf(w.w, "\r")
}
