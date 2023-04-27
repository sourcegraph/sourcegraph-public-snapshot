package log

import (
	"encoding/json"
	"io"
	"time"

	batcheslib "github.com/sourcegraph/sourcegraph/lib/batches"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Logger is a simple wrapper around an io.Writer that writes batcheslib.LogEvents.
type Logger struct {
	Writer io.Writer
}

// WriteEvent writes a batcheslib.LogEvent to the underlying io.Writer.
func (l *Logger) WriteEvent(operation batcheslib.LogEventOperation, status batcheslib.LogEventStatus, metadata any) error {
	e := batcheslib.LogEvent{Operation: operation, Status: status, Metadata: metadata}
	e.Timestamp = time.Now().UTC().Truncate(time.Millisecond)
	if err := json.NewEncoder(l.Writer).Encode(e); err != nil {
		return errors.Wrap(err, "failed to encode event")
	}
	return nil
}
