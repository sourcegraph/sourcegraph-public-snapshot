package worker

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"sourcegraph.com/sourcegraph/sourcegraph/api/sourcegraph"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
)

var (
	// logw is the io.Writer that all log output should go to.
	logw io.Writer = os.Stderr
)

func init() {
	if logFilename := os.Getenv("SG_LOG_FILE"); logFilename != "" {
		f, err := os.OpenFile(logFilename, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
		if err != nil {
			log.Fatal(err)
		}
		logw = io.MultiWriter(f, os.Stderr)
		log.SetOutput(logw)
	}
}

type logger struct {
	*log.Logger
	io.Writer
	Destination string
	mutex       sync.Mutex
	c           []io.Closer
}

type LogWriter interface {
	io.Writer
	io.Closer
}

func (x *logger) String() string { return x.Destination }

func (x *logger) Write(p []byte) (n int, err error) {
	x.mutex.Lock()
	defer x.mutex.Unlock()
	return x.Writer.Write(p)
}

func (x *logger) Close() error {
	x.mutex.Lock()
	defer x.mutex.Unlock()
	if len(x.c) == 0 {
		return nil
	}
	for _, c := range x.c {
		err := c.Close()
		if err != nil {
			return err
		}
	}
	x.c = nil
	return nil
}

var logURLForTagFormat = os.Getenv("SG_LOG_URL_FOR_TAG")

func logURLForTag(tag string) string {
	if logURLForTagFormat == "" {
		return ""
	}
	return fmt.Sprintf(logURLForTagFormat, url.QueryEscape(tag))
}

func createPath(path string) (LogWriter, error) {
	if dir := filepath.Dir(path); dir != "." {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return nil, err
		}
	}
	return os.Create(path)
}

// newLogger creates a new logger for a build task that writes to
// Papertrail (remote syslog) and/or local files. Call CloseLogs
// before the program exits.
func newLogger(task sourcegraph.TaskSpec) *logger {
	tag := task.IDString()

	var dests []string
	var ws []io.Writer
	var cs []io.Closer

	pw, err := newPapertrailLogger(tag)
	if err == nil {
		ws = append(ws, pw)
		cs = append(cs, pw)
		dests = append(dests, logURLForTag(tag))
	} else if usePapertrail {
		log.Printf("Warning: Remote syslog logging disabled because remote syslog host connection failed: %s.", err)
	}

	if err := os.MkdirAll(conf.BuildLogDir, 0700); err == nil {
		logFile := filepath.Join(conf.BuildLogDir, tag+".log")
		fw, err := createPath(logFile)
		if err != nil {
			log.Printf("Warning: file logging disabled because log file could not be opened: %s.", err)
		}

		if fw != nil {
			ws = append(ws, fw)
			cs = append(cs, fw)
			dests = append(dests, logFile)
		}
	} else {
		log.Printf("Warning: file logging disabled because log directory could not be created: %s.", err)
	}

	mw := io.MultiWriter(ws...)

	l := &logger{
		Logger:      log.New(mw, "", 0),
		Writer:      mw,
		Destination: strings.Join(dests, " and "),
		c:           cs,
	}

	closersMu.Lock()
	defer closersMu.Unlock()
	closers = append(closers, l)

	return l
}

var (
	closers   []io.Closer
	closersMu sync.Mutex
)

// CloseLogs ensures that all log lines have been written to the log
// files and destinations (e.g., remote syslog) before returning.
func CloseLogs() {
	closersMu.Lock()
	defer closersMu.Unlock()
	var w sync.WaitGroup
	for _, c := range closers {
		w.Add(1)
		go func(c io.Closer) {
			defer w.Done()
			if err := c.Close(); err != nil {
				log.Printf("Error flushing log: %s.", err)
			}
		}(c)
	}
	w.Wait()
}
