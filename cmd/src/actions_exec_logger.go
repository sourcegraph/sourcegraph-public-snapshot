package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
	"github.com/pkg/errors"
)

var (
	boldRed   = color.New(color.Bold, color.FgRed)
	boldGreen = color.New(color.Bold, color.FgGreen)
	green     = color.New(color.FgGreen)
	yellow    = color.New(color.FgYellow)
	grey      = color.New(color.FgHiBlack)
)

type actionLogger struct {
	verbose  bool
	keepLogs bool

	highlight func(a ...interface{}) string

	logFiles   map[string]*os.File
	logWriters map[string]io.Writer
	mu         sync.Mutex
}

func newActionLogger(verbose, keepLogs bool) *actionLogger {
	useColor := isatty.IsTerminal(os.Stderr.Fd()) || isatty.IsCygwinTerminal(os.Stderr.Fd())
	if useColor {
		color.NoColor = false
	}
	return &actionLogger{
		verbose:    verbose,
		keepLogs:   keepLogs,
		highlight:  color.New(color.Bold, color.FgGreen).SprintFunc(),
		logFiles:   map[string]*os.File{},
		logWriters: map[string]io.Writer{},
	}
}

func (a *actionLogger) Warnf(format string, args ...interface{}) {
	if a.verbose {
		fmt.Fprintf(os.Stderr, color.YellowString("WARNING: ")+format, args...)
	}
}

func (a *actionLogger) Infof(format string, args ...interface{}) {
	if a.verbose {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func (a *actionLogger) RepoCacheHit(repo ActionRepo, patchProduced bool) {
	if a.verbose {
		if patchProduced {
			fmt.Fprintf(os.Stderr, "%s -> Cached result with patch found.\n", boldGreen.Sprint(repo.Name))
			return
		}

		fmt.Fprintf(os.Stderr, "%s -> Cached result without patch found.\n", green.Sprint(repo.Name))
	}
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

	if a.verbose {
		fmt.Fprintf(os.Stderr, "%s -> Enqueued. Logfile created at %s\n", grey.Sprint(repo.Name), logFile.Name())
	}

	return logFile.Name(), nil
}

func (a *actionLogger) RepoWriter(repoName string) (io.Writer, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w, ok := a.logWriters[repoName]
	return w, ok
}

func (a *actionLogger) write(repoName string, c *color.Color, format string, args ...interface{}) {
	if w, ok := a.RepoWriter(repoName); ok {
		fmt.Fprintf(w, format, args...)
	}

	if a.verbose {
		format = fmt.Sprintf("%s -> %s", c.Sprint(repoName), format)
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func (a *actionLogger) RepoFinished(repoName string, patchProduced bool, actionErr error) error {
	if actionErr != nil {
		a.write(repoName, boldRed, "%s\n", actionErr)
	} else if patchProduced {
		a.write(repoName, boldGreen, "Patch generated.\n")
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	f, ok := a.logFiles[repoName]
	if !ok {
		return nil
	}

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
	a.write(repoName, yellow, "Step %d: command %v\n", step, args)
}

func (a *actionLogger) CommandStepErrored(repoName string, step int, err error) {
	a.write(repoName, boldRed, "Step %d: error: %s.\n", step, err)
}

func (a *actionLogger) CommandStepDone(repoName string, step int) {
	a.write(repoName, yellow, "Step %d: done.\n", step)
}

func (a *actionLogger) DockerStepStarted(repoName string, step int, dockerfile, image string) {
	var fromDockerfile string
	if dockerfile != "" {
		fromDockerfile = " (built from inline Dockerfile)"
	}
	a.write(repoName, yellow, "Step %d: docker run %v%s\n", step, image, fromDockerfile)
}

func (a *actionLogger) DockerStepErrored(repoName string, step int, err error, elapsed time.Duration) {
	a.write(repoName, boldRed, "Step %d: error: %s. (%s)\n", step, err, elapsed)
}

func (a *actionLogger) DockerStepDone(repoName string, step int, elapsed time.Duration) {
	a.write(repoName, yellow, "Step %d: done. (%s)\n", step, elapsed)
}
