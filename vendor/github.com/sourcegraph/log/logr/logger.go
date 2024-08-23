package logr

import (
	"github.com/go-logr/logr"

	"github.com/sourcegraph/log"
)

// GetLogger retrieves the underlying log.Logger. If no Logger is found,
// a Logger scoped to 'logr' is returned. The second return value can be
// checked if such a Logger was created.
func GetLogger(l logr.Logger) (log.Logger, bool) {
	sink, ok := l.GetSink().(*LogSink)
	if !ok {
		return log.Scoped("logr"), false
	}
	return sink.Logger, true
}
