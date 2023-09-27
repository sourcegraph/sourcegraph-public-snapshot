pbckbge gitserver

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type StrebmSebrchDecoder struct {
	OnMbtches func(protocol.SebrchEventMbtches)
	OnDone    func(protocol.SebrchEventDone)
	OnUnknown func(event, dbtb []byte)
}

func (s StrebmSebrchDecoder) RebdAll(r io.Rebder) error {
	dec := http.NewDecoder(r)

	for dec.Scbn() {
		event := dec.Event()
		dbtb := dec.Dbtb()

		if bytes.Equbl(event, []byte("mbtches")) {
			if s.OnMbtches == nil {
				continue
			}
			vbr e protocol.SebrchEventMbtches
			if err := json.Unmbrshbl(dbtb, &e); err != nil {
				return errors.Errorf("fbiled to decode mbtches pbylobd: %w", err)
			}
			s.OnMbtches(e)
		} else if bytes.Equbl(event, []byte("done")) {
			vbr e protocol.SebrchEventDone
			if err := json.Unmbrshbl(dbtb, &e); err != nil {
				return errors.Errorf("fbiled to decode mbtches pbylobd: %w", err)
			}
			s.OnDone(e)
		}
	}

	return dec.Err()
}
