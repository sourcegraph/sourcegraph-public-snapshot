package campaigns

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

type ActionLogger struct {
	verbose  bool
	keepLogs bool

	progress *progress
	out      io.WriteCloser

	mu         sync.Mutex
	logFiles   map[string]*os.File
	logWriters map[string]io.Writer
}

func NewActionLogger(verbose, keepLogs bool) *ActionLogger {
	useColor := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
	if useColor {
		color.NoColor = false
	}

	progress := new(progress)

	return &ActionLogger{
		verbose:  verbose,
		keepLogs: keepLogs,
		progress: progress,
		out: &progressWriter{
			p: progress,
			w: os.Stderr,
		},
		logFiles:   map[string]*os.File{},
		logWriters: map[string]io.Writer{},
	}
}

func (a *ActionLogger) Start(totalSteps int) {
	a.progress.SetTotalSteps(int64(totalSteps))
}

func (a *ActionLogger) Infof(format string, args ...interface{}) {
	if a.verbose {
		a.log("", grey, format, args...)
	}
}

func (a *ActionLogger) Warnf(format string, args ...interface{}) {
	if a.verbose {
		a.log("", yellow, "WARNING: "+format, args...)
	}
}

func (a *ActionLogger) ActionFailed(err error, patches []PatchInput) {
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

func (a *ActionLogger) ActionSuccess(patches []PatchInput) {
	a.out.Close()
	fmt.Fprintln(os.Stderr)
	format := "✔  Action produced %d patches."
	hiGreen.Fprintf(os.Stderr, format, len(patches))
}

func (a *ActionLogger) RepoCacheHit(repo ActionRepo, stepCount int, patchProduced bool) {
	a.progress.IncStepsComplete(int64(stepCount))
	if patchProduced {
		a.progress.IncPatchCount()
		a.log(repo.Name, boldGreen, "Cached result found: using cached diff.\n")
		return
	}
	a.log(repo.Name, grey, "Cached result found: no diff produced for this repository.\n")
}

func (a *ActionLogger) AddRepo(repo ActionRepo) (string, error) {
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

func (a *ActionLogger) RepoWriter(repoName string) (io.Writer, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w, ok := a.logWriters[repoName]
	return w, ok
}

func (a *ActionLogger) InfoPipe(prefix string) io.Writer {
	stdoutPrefix := fmt.Sprintf("%s -> [STDOUT]: ", yellow.Sprint(prefix))
	stderr := textio.NewPrefixWriter(os.Stderr, stdoutPrefix)
	return io.Writer(stderr)
}

func (a *ActionLogger) ErrorPipe(prefix string) io.Writer {
	stderrPrefix := fmt.Sprintf("%s -> [STDERR]: ", yellow.Sprint(prefix))
	stderr := textio.NewPrefixWriter(os.Stderr, stderrPrefix)
	return io.Writer(stderr)
}

func (a *ActionLogger) RepoStdoutStderr(repoName string) (io.Writer, io.Writer, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w, ok := a.logWriters[repoName]

	stderrPrefix := fmt.Sprintf("%s -> [STDERR]: ", yellow.Sprint(repoName))
	stderr := textio.NewPrefixWriter(a.out, stderrPrefix)

	stdoutPrefix := fmt.Sprintf("%s -> [STDOUT]: ", yellow.Sprint(repoName))
	stdout := textio.NewPrefixWriter(a.out, stdoutPrefix)

	return io.MultiWriter(stdout, w), io.MultiWriter(stderr, w), ok
}

func (a *ActionLogger) RepoFinished(repoName string, patchProduced bool, actionErr error) error {
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

func (a *ActionLogger) RepoStarted(repoName, rev string, steps []*ActionStep) {
	a.write(repoName, yellow, "Starting action @ %s (%d steps)\n", rev, len(steps))
}

func (a *ActionLogger) CommandStepStarted(repoName string, step int, args []string) {
	a.write(repoName, yellow, "%s command %v\n", boldBlack.Sprintf("[Step %d]", step), args)
}

func (a *ActionLogger) CommandStepErrored(repoName string, step int, err error) {
	a.progress.IncStepsComplete(1)
	a.progress.IncStepsFailed()
	a.write(repoName, boldRed, "%s %s.\n", boldBlack.Sprintf("[Step %d]", step), err)
}

func (a *ActionLogger) CommandStepDone(repoName string, step int) {
	a.progress.IncStepsComplete(1)
	a.write(repoName, yellow, "%s Done.\n", boldBlack.Sprintf("[Step %d]", step))
}

func (a *ActionLogger) DockerStepStarted(repoName string, step int, image string) {
	a.write(repoName, yellow, "%s docker run %s\n", boldBlack.Sprintf("[Step %d]", step), image)
}

func (a *ActionLogger) DockerStepErrored(repoName string, step int, err error, elapsed time.Duration) {
	a.progress.IncStepsComplete(1)
	a.progress.IncStepsFailed()
	a.write(repoName, boldRed, "%s %s. (%s)\n", boldBlack.Sprintf("[Step %d]", step), err, elapsed)
}

func (a *ActionLogger) DockerStepDone(repoName string, step int, elapsed time.Duration) {
	a.progress.IncStepsComplete(1)
	a.write(repoName, yellow, "%s Done. (%s)\n", boldBlack.Sprintf("[Step %d]", step), elapsed)
}

func (a *ActionLogger) RepoMatches(repoCount int, skipped, unsupported []string) {
	for _, r := range skipped {
		a.Infof("Skipping repository %s because we couldn't determine default branch.\n", r)
	}
	unsupportedCount := len(unsupported)
	var matchesStr string
	if repoCount == 1 {
		matchesStr = fmt.Sprintf("%d repository matches the scopeQuery.", repoCount)
	} else {
		var warnStr string
		if repoCount == 0 {
			warnStr = "WARNING: "
		}
		matchesStr = fmt.Sprintf("%s%d repositories match the scopeQuery.", warnStr, repoCount)
	}
	if unsupportedCount > 0 {
		matchesStr += fmt.Sprintf("\n\n%d repositories were filtered out because they are on a codehost not supported by campaigns. (use -include-unsupported to generate patches for them anyway):\n", unsupportedCount)
		for i, repo := range unsupported {
			matchesStr += color.HiYellowString("- %s\n", repo)
			if i == 10 {
				matchesStr += fmt.Sprintf("and %d more.\n", unsupportedCount-10)
				break
			}
		}
	}
	color := yellow
	if repoCount > 0 {
		color = boldGreen
	}
	a.write("", color, "%s\n\n", matchesStr)
}

// write writes to the RepoWriter associated with the given repoName and logs the message using the log method.
func (a *ActionLogger) write(repoName string, c *color.Color, format string, args ...interface{}) {
	if w, ok := a.RepoWriter(repoName); ok {
		fmt.Fprintf(w, format, args...)
	}
	a.log(repoName, c, format, args...)
}

// log logs only to stderr, it does not log to our repoWriters.
func (a *ActionLogger) log(repoName string, c *color.Color, format string, args ...interface{}) {
	if len(repoName) > 0 {
		format = fmt.Sprintf("%s -> %s", c.Sprint(repoName), format)
	}
	fmt.Fprintf(a.out, c.Sprintf(format, args...))
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

	if !bytes.HasSuffix(data, []byte("\n")) && !bytes.HasSuffix(data, []byte("\n\x1b[0m")) {
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
	progressText := fmt.Sprintf("[%s] Steps: %d/%d (%s, %s)", bar, w.p.StepsComplete(), w.p.TotalSteps(), boldRed.Sprintf("%d failed", w.p.TotalStepsFailed()), hiGreen.Sprintf("%d patches", w.p.PatchCount()))
	fmt.Fprint(w.w, progressText)
	w.shouldClear = true
	w.progressLogLength = len(progressText)
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
	fmt.Fprint(w.w, "\r")
	fmt.Fprint(w.w, strings.Repeat(" ", w.progressLogLength))
	fmt.Fprint(w.w, "\r")
}
