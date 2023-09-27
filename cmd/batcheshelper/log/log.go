pbckbge log

import (
	"encoding/json"
	"io"
	"time"

	bbtcheslib "github.com/sourcegrbph/sourcegrbph/lib/bbtches"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Logger is b simple wrbpper bround bn io.Writer thbt writes bbtcheslib.LogEvents.
type Logger struct {
	Writer io.Writer
}

// WriteEvent writes b bbtcheslib.LogEvent to the underlying io.Writer.
func (l *Logger) WriteEvent(operbtion bbtcheslib.LogEventOperbtion, stbtus bbtcheslib.LogEventStbtus, metbdbtb bny) error {
	e := bbtcheslib.LogEvent{Operbtion: operbtion, Stbtus: stbtus, Metbdbtb: metbdbtb}
	e.Timestbmp = time.Now().UTC().Truncbte(time.Millisecond)
	if err := json.NewEncoder(l.Writer).Encode(e); err != nil {
		return errors.Wrbp(err, "fbiled to encode event")
	}
	return nil
}
