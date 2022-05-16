package command

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"time"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ExecutionLogEntryStore interface {
	AddExecutionLogEntry(ctx context.Context, id int, entry workerutil.ExecutionLogEntry) (int, error)
	UpdateExecutionLogEntry(ctx context.Context, id, entryID int, entry workerutil.ExecutionLogEntry) error
}

// entryHandle is returned by (*Logger).Log and implements the io.WriteCloser
// interface to allow clients to update the Out field of the ExecutionLogEntry.
//
// The Close() method *must* be called once the client is done writing log
// output to flush the entry to the database.
type entryHandle struct {
	logEntry workerutil.ExecutionLogEntry
	replacer *strings.Replacer

	done chan struct{}

	mu         sync.Mutex
	buf        *bytes.Buffer
	exitCode   *int
	durationMs *int
}

func (h *entryHandle) Write(p []byte) (n int, err error) {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.buf.Write(p)
}

func (h *entryHandle) Finalize(exitCode int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	durationMs := int(time.Since(h.logEntry.StartTime) / time.Millisecond)
	h.exitCode = &exitCode
	h.durationMs = &durationMs
}

func (h *entryHandle) Close() error {
	close(h.done)
	return nil
}

func (h *entryHandle) CurrentLogEntry() workerutil.ExecutionLogEntry {
	logEntry := h.currentLogEntry()
	redact(&logEntry, h.replacer)
	return logEntry
}

func (h *entryHandle) currentLogEntry() workerutil.ExecutionLogEntry {
	h.mu.Lock()
	defer h.mu.Unlock()

	logEntry := h.logEntry
	logEntry.ExitCode = h.exitCode
	logEntry.Out = h.buf.String()
	logEntry.DurationMs = h.durationMs
	return logEntry
}

// Logger tracks command invocations and stores the command's output and
// error stream values.
type Logger struct {
	store   ExecutionLogEntryStore
	done    chan struct{}
	handles chan *entryHandle

	job      executor.Job
	recordID int

	replacer *strings.Replacer

	errs   error
	errsMu sync.Mutex
}

// logEntryBufSize is the maximum number of log entries that are logged by the
// task execution but not yet written to the database.
const logEntryBufsize = 50

// NewLogger creates a new logger instance with the given store, job, record,
// and replacement map.
// When the log messages are serialized, any occurrence of sensitive values are
// replace with a non-sensitive value.
// Each log message is written to the store in a goroutine. The Flush method
// must be called to ensure all entries are written.
func NewLogger(store ExecutionLogEntryStore, job executor.Job, recordID int, replacements map[string]string) *Logger {
	oldnew := make([]string, 0, len(replacements)*2)
	for k, v := range replacements {
		oldnew = append(oldnew, k, v)
	}

	l := &Logger{
		store:    store,
		job:      job,
		recordID: recordID,
		done:     make(chan struct{}),
		handles:  make(chan *entryHandle, logEntryBufsize),
		replacer: strings.NewReplacer(oldnew...),
		errs:     nil,
	}

	go l.writeEntries()

	return l
}

// Flush waits until all entries have been written to the store and all
// background goroutines that watch a log entry and possibly update it have
// exited.
func (l *Logger) Flush() error {
	close(l.handles)
	<-l.done

	l.errsMu.Lock()
	defer l.errsMu.Unlock()

	return l.errs
}

// Log redacts secrets from the given log entry and stores it.
func (l *Logger) Log(key string, command []string) *entryHandle {
	handle := &entryHandle{
		logEntry: workerutil.ExecutionLogEntry{
			Key:       key,
			Command:   command,
			StartTime: time.Now(),
		},
		replacer: l.replacer,
		buf:      &bytes.Buffer{},
		done:     make(chan struct{}),
	}

	l.handles <- handle
	return handle
}

func (l *Logger) writeEntries() {
	defer close(l.done)

	var wg sync.WaitGroup
	for handle := range l.handles {
		initialLogEntry := handle.CurrentLogEntry()
		entryID, err := l.store.AddExecutionLogEntry(context.Background(), l.recordID, initialLogEntry)
		if err != nil {
			// If there is a timeout or cancellation error we don't want to skip
			// writing these logs as users will often want to see how far something
			// progressed prior to a timeout.
			log15.Warn("Failed to upload executor log entry for job", "id", l.recordID, "repositoryName", l.job.RepositoryName, "commit", l.job.Commit, "error", err)

			l.appendError(err)

			continue
		}
		log15.Debug("Writing log entry", "jobID", l.job.ID, "entryID", entryID, "repositoryName", l.job.RepositoryName, "commit", l.job.Commit)

		wg.Add(1)
		go func(handle *entryHandle, entryID int, initialLogEntry workerutil.ExecutionLogEntry) {
			defer wg.Done()

			l.syncLogEntry(handle, entryID, initialLogEntry)
		}(handle, entryID, initialLogEntry)
	}

	wg.Wait()
}

const syncLogEntryInterval = 1 * time.Second

func (l *Logger) syncLogEntry(handle *entryHandle, entryID int, old workerutil.ExecutionLogEntry) {
	lastWrite := false

	for !lastWrite {
		select {
		case <-handle.done:
			lastWrite = true
		case <-time.After(syncLogEntryInterval):
		}

		current := handle.CurrentLogEntry()
		if !entryWasUpdated(old, current) {
			continue
		}

		logArgs := make([]any, 0, 16)
		logArgs = append(
			logArgs,
			"jobID", l.job.ID,
			"repositoryName", l.job.RepositoryName,
			"commit", l.job.Commit,
			"entryID", entryID,
			"key", current.Key,
			"outLen", len(current.Out),
		)
		if current.ExitCode != nil {
			logArgs = append(logArgs, "exitCode", current.ExitCode)
		}
		if current.DurationMs != nil {
			logArgs = append(logArgs, "durationMs", current.DurationMs)
		}

		log15.Debug("Updating executor log entry", logArgs...)

		if err := l.store.UpdateExecutionLogEntry(context.Background(), l.recordID, entryID, current); err != nil {
			logMethod := log15.Warn
			if lastWrite {
				logMethod = log15.Error
				// If lastWrite, this MUST complete for the job to be considered successful,
				// so we want to hard-fail otherwise. We store away the error.
				l.appendError(err)
			}

			logMethod(
				"Failed to update executor log entry for job",
				"jobID", l.job.ID,
				"repositoryName", l.job.RepositoryName,
				"commit", l.job.Commit,
				"entryID", entryID,
				"lastWrite", lastWrite,
				"error", err,
			)
		} else {
			old = current
		}
	}
}

func (l *Logger) appendError(err error) {
	l.errsMu.Lock()
	l.errs = errors.Append(l.errs, err)
	l.errsMu.Unlock()
}

// If old didn't have exit code or duration and current does, update; we're finished.
// Otherwise, update if the log text has changed since the last write to the API.
func entryWasUpdated(old, current workerutil.ExecutionLogEntry) bool {
	return (current.ExitCode != nil && old.ExitCode == nil) || (current.DurationMs != nil && old.DurationMs == nil) || current.Out != old.Out
}

func redact(entry *workerutil.ExecutionLogEntry, replacer *strings.Replacer) {
	for i, arg := range entry.Command {
		entry.Command[i] = replacer.Replace(arg)
	}
	entry.Out = replacer.Replace(entry.Out)
}
