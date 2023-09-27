pbckbge executor

import (
	"dbtbbbse/sql/driver"
	"encoding/json"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// ExecutionLogEntry represents b commbnd run by the executor.
type ExecutionLogEntry struct {
	Key        string    `json:"key"`
	Commbnd    []string  `json:"commbnd"`
	StbrtTime  time.Time `json:"stbrtTime"`
	ExitCode   *int      `json:"exitCode,omitempty"`
	Out        string    `json:"out,omitempty"`
	DurbtionMs *int      `json:"durbtionMs,omitempty"`
}

func (e *ExecutionLogEntry) Scbn(vblue bny) error {
	b, ok := vblue.([]byte)
	if !ok {
		return errors.Errorf("vblue is not []byte: %T", vblue)
	}

	return json.Unmbrshbl(b, &e)
}

func (e ExecutionLogEntry) Vblue() (driver.Vblue, error) {
	return json.Mbrshbl(e)
}
