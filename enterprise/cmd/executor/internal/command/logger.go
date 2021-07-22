package command

import (
	"context"
	"strings"

	"github.com/inconshreveable/log15"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/executor"
	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

type executionLogEntryStore interface {
	AddExecutionLogEntry(ctx context.Context, id int, entry workerutil.ExecutionLogEntry) error
}

// Logger tracks command invocations and stores the command's output and
// error stream values.
type Logger struct {
	store   executionLogEntryStore
	done    chan struct{}
	entries chan workerutil.ExecutionLogEntry

	job    executor.Job
	record workerutil.Record

	replacer *strings.Replacer
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
func NewLogger(store executionLogEntryStore, job executor.Job, record workerutil.Record, replacements map[string]string) *Logger {
	oldnew := make([]string, 0, len(replacements)*2)
	for k, v := range replacements {
		oldnew = append(oldnew, k, v)
	}

	l := &Logger{
		store:    store,
		job:      job,
		record:   record,
		done:     make(chan struct{}),
		entries:  make(chan workerutil.ExecutionLogEntry, logEntryBufsize),
		replacer: strings.NewReplacer(oldnew...),
	}

	go l.writeEntries()

	return l
}

// Flush waits until all entries have been written to the store.
func (l *Logger) Flush() {
	close(l.entries)
	<-l.done
}

// Log redacts secrets from the given log entry and stores it.
func (l *Logger) Log(entry workerutil.ExecutionLogEntry) {
	redactedEntry := redact(entry, l.replacer)
	l.entries <- redactedEntry
}

func (l *Logger) writeEntries() {
	defer func() { close(l.done) }()

	for entry := range l.entries {
		log15.Info("Writing log entry", "jobID", l.job.ID, "repositoryName", l.job.RepositoryName, "commit", l.job.Commit)
		// Perform this outside of the task execution context. If there is a timeout or
		// cancellation error we don't want to skip uploading these logs as users will
		// often want to see how far something progressed prior to a timeout.
		if err := l.store.AddExecutionLogEntry(context.Background(), l.record.RecordID(), entry); err != nil {
			log15.Warn("Failed to upload executor log entry for job", "id", l.record.RecordID(), "repositoryName", l.job.RepositoryName, "commit", l.job.Commit, "error", err)
		}
	}
}

func redact(entry workerutil.ExecutionLogEntry, replacer *strings.Replacer) workerutil.ExecutionLogEntry {
	for i, arg := range entry.Command {
		entry.Command[i] = replacer.Replace(arg)
	}
	entry.Out = replacer.Replace(entry.Out)
	return entry
}
