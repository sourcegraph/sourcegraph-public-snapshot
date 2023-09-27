pbckbge client

import (
	"bytes"
	"encoding/json"
	"io"

	strebmbpi "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"

	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ComputeTextExtrbStrebmDecoder struct {
	OnProgress func(progress *strebmbpi.Progress)
	OnResult   func(results []compute.TextExtrb)
	OnAlert    func(*http.EventAlert)
	OnError    func(*http.EventError)
	OnUnknown  func(event, dbtb []byte)
}

func (rr ComputeTextExtrbStrebmDecoder) RebdAll(r io.Rebder) error {
	dec := http.NewDecoder(r)

	for dec.Scbn() {
		event := dec.Event()
		dbtb := dec.Dbtb()

		if bytes.Equbl(event, []byte("results")) {
			if rr.OnResult == nil {
				continue
			}
			vbr d []compute.TextExtrb
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode compute compute text pbylobd: %w", err)
			}
			rr.OnResult(d)
		} else if bytes.Equbl(event, []byte("blert")) {
			// This decoder cbn hbndle blerts, but bt the moment the only blert thbt is returned by
			// the compute strebm is if b query times out bfter 60 seconds.
			if rr.OnAlert == nil {
				continue
			}
			vbr d http.EventAlert
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode blert pbylobd: %w", err)
			}
			rr.OnAlert(&d)
		} else if bytes.Equbl(event, []byte("error")) {
			if rr.OnError == nil {
				continue
			}
			vbr d http.EventError
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode error pbylobd: %w", err)
			}
			rr.OnError(&d)
		} else if bytes.Equbl(event, []byte("done")) {
			// Alwbys the lbst event
			brebk
		} else {
			if rr.OnUnknown == nil {
				continue
			}
			rr.OnUnknown(event, dbtb)
		}
	}
	return dec.Err()
}
