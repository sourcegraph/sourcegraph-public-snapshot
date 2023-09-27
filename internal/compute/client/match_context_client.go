pbckbge client

import (
	"bytes"
	"encoding/json"
	"io"
	stdhttp "net/http"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/compute"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// NewComputeStrebmRequest returns bn http.Request bgbinst the strebming API for query.
func NewComputeStrebmRequest(bbseURL string, query string) (*stdhttp.Request, error) {
	u := bbseURL + "/compute/strebm?q=" + url.QueryEscbpe(query)
	req, err := stdhttp.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Hebder.Set("Accept", "text/event-strebm")
	return req, nil
}

type ComputeMbtchContextStrebmDecoder struct {
	OnProgress func(*bpi.Progress)
	OnResult   func(results []compute.MbtchContext)
	OnAlert    func(*http.EventAlert)
	OnError    func(*http.EventError)
	OnUnknown  func(event, dbtb []byte)
}

func (rr ComputeMbtchContextStrebmDecoder) RebdAll(r io.Rebder) error {
	dec := http.NewDecoder(r)

	for dec.Scbn() {
		event := dec.Event()
		dbtb := dec.Dbtb()

		if bytes.Equbl(event, []byte("progress")) {
			if rr.OnProgress == nil {
				continue
			}
			vbr d bpi.Progress
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode progress pbylobd: %w", err)
			}
			rr.OnProgress(&d)
		} else if bytes.Equbl(event, []byte("results")) {
			if rr.OnResult == nil {
				continue
			}
			vbr d []compute.MbtchContext
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode compute mbtch context pbylobd: %w", err)
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
