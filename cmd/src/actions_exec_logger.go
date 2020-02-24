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
		fmt.Fprintf(os.Stderr, "WARNING: "+format, args...)
	}
}

func (a *actionLogger) Infof(format string, args ...interface{}) {
	if a.verbose {
		fmt.Fprintf(os.Stderr, format, args...)
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
		fmt.Fprintf(os.Stderr, "%s -> Logfile created at %s\n", a.highlight(repo.Name), logFile.Name())
	}

	return logFile.Name(), nil
}

func (a *actionLogger) RepoWriter(repoName string) (io.Writer, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	w, ok := a.logWriters[repoName]
	return w, ok
}

func (a *actionLogger) write(repoName, format string, args ...interface{}) {
	if w, ok := a.RepoWriter(repoName); ok {
		fmt.Fprintf(w, format, args...)
	}

	if a.verbose {
		format = fmt.Sprintf("%s -> %s", a.highlight(repoName), format)
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

func (a *actionLogger) RepoFinished(repoName string, actionErr error) error {
	if actionErr != nil {
		a.write(repoName, "%s: ERROR: %s\n", repoName, actionErr)
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
	a.write(repoName, "Starting action @ %s (%d steps)\n", rev, len(steps))
}

func (a *actionLogger) CommandStepStarted(repoName string, step int, args []string) {
	a.write(repoName, "Step %d: command %v\n", step, args)
}

func (a *actionLogger) CommandStepErrored(repoName string, step int, err error) {
	a.write(repoName, "Step %d: error: %s.\n", step, err)
}

func (a *actionLogger) CommandStepDone(repoName string, step int) {
	a.write(repoName, "Step %d: done.\n", step)
}

func (a *actionLogger) DockerStepStarted(repoName string, step int, dockerfile, image string) {
	var fromDockerfile string
	if dockerfile != "" {
		fromDockerfile = " (built from inline Dockerfile)"
	}
	a.write(repoName, "Step %d: docker run %v%s\n", step, image, fromDockerfile)
}

func (a *actionLogger) DockerStepErrored(repoName string, step int, err error, elapsed time.Duration) {
	a.write(repoName, "Step %d: error: %s. (%s)\n", step, err, elapsed)
}

func (a *actionLogger) DockerStepDone(repoName string, step int, elapsed time.Duration) {
	a.write(repoName, "Step %d: done. (%s)\n", step, elapsed)
}
