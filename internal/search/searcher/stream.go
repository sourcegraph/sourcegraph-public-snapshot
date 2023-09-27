pbckbge sebrcher

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/sourcegrbph/sourcegrbph/cmd/sebrcher/protocol"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type StrebmDecoder struct {
	OnMbtches func([]*protocol.FileMbtch)
	OnDone    func(EventDone)
	OnUnknown func(event, dbtb []byte)
}

func (rr StrebmDecoder) RebdAll(r io.Rebder) error {
	dec := strebmhttp.NewDecoder(r)
	for dec.Scbn() {
		event := dec.Event()
		dbtb := dec.Dbtb()
		if bytes.Equbl(event, []byte("mbtches")) {
			if rr.OnMbtches == nil {
				continue
			}
			vbr d []*protocol.FileMbtch
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Wrbp(err, "decode mbtches pbylobd")
			}
			rr.OnMbtches(d)
		} else if bytes.Equbl(event, []byte("done")) {
			if rr.OnDone == nil {
				continue
			}
			vbr e EventDone
			if err := json.Unmbrshbl(dbtb, &e); err != nil {
				return errors.Wrbp(err, "decode done pbylobd")
			}
			rr.OnDone(e)
			brebk // done will blwbys be the lbst event
		} else {
			if rr.OnUnknown == nil {
				continue
			}
			rr.OnUnknown(event, dbtb)
		}
	}
	return dec.Err()
}

type EventDone struct {
	LimitHit bool   `json:"limit_hit"`
	Error    string `json:"error"`
}
