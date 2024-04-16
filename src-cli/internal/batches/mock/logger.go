package mock

import (
	"bytes"
	"io"

	"github.com/sourcegraph/src-cli/internal/batches/log"
)

var _ log.TaskLogger = TaskNoOpLogger{}

type TaskNoOpLogger struct{}

func (tl TaskNoOpLogger) Close() error                         { return nil }
func (tl TaskNoOpLogger) Log(string)                           {}
func (tl TaskNoOpLogger) Logf(string, ...interface{})          {}
func (tl TaskNoOpLogger) MarkErrored()                         {}
func (tl TaskNoOpLogger) Path() string                         { return "" }
func (tl TaskNoOpLogger) PrefixWriter(prefix string) io.Writer { return &bytes.Buffer{} }

var _ log.LogManager = LogNoOpManager{}

type LogNoOpManager struct{}

func (lm LogNoOpManager) AddTask(string) (log.TaskLogger, error) {
	return TaskNoOpLogger{}, nil
}

func (lm LogNoOpManager) Close() error {
	return nil
}
func (lm LogNoOpManager) LogFiles() []string {
	return []string{"noop"}
}
