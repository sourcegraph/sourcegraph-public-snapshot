package log

import "io"

type LogManager interface {
	AddTask(string) (TaskLogger, error)
	Close() error
	LogFiles() []string
}

type TaskLogger interface {
	Close() error
	Log(string)
	Logf(string, ...interface{})
	MarkErrored()
	Path() string
	PrefixWriter(prefix string) io.Writer
}
