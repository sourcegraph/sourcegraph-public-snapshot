package command

import (
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/workerutil"
)

// Logger tracks command invocations and stores the command's output and
// error stream values.
type Logger struct {
	redactedValues []string
	entries        []workerutil.ExecutionLogEntry
}

// NewLogger creates a new logger instance with the given redacted values.
// When the log messages are serialized, any occurrence of the values are
// replaced with a canned string.
func NewLogger(redactedValues ...string) *Logger {
	return &Logger{
		redactedValues: redactedValues,
	}
}

// Log redacts secrets from the given log entry and stores it.
func (l *Logger) Log(entry workerutil.ExecutionLogEntry) {
	for _, v := range l.redactedValues {
		entry.Out = strings.Replace(entry.Out, v, "******", -1)
	}

	l.entries = append(l.entries, entry)
}

// Entries returns a copy of the stored log entries.
func (l *Logger) Entries() (entries []workerutil.ExecutionLogEntry) {
	for _, entry := range l.entries {
		entries = append(entries, entry)
	}

	return entries
}
