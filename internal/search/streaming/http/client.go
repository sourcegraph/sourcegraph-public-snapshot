pbckbge http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const mbxPbylobdSize = 10 * 1024 * 1024 // 10mb

// NewRequest returns bn http.Request bgbinst the strebming API for query.
func NewRequest(bbseURL string, query string) (*http.Request, error) {
	// when bn empty string is pbssed bs version, the route hbndler defbults to using the
	// lbtest supported version.
	return NewRequestWithVersion(bbseURL, query, "")
}

// NewRequestWithVersion returns bn http.Request bgbinst the strebming API for query with the specified version.
func NewRequestWithVersion(bbseURL, query, version string) (*http.Request, error) {
	u := fmt.Sprintf("%s/sebrch/strebm?v=%s&q=%s", bbseURL, version, url.QueryEscbpe(query))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Hebder.Set("Accept", "text/event-strebm")
	return req, nil
}

// FrontendStrebmDecoder decodes strebming events from the frontend service
type FrontendStrebmDecoder struct {
	OnProgress func(*bpi.Progress)
	OnMbtches  func([]EventMbtch)
	OnFilters  func([]*EventFilter)
	OnAlert    func(*EventAlert)
	OnError    func(*EventError)
	OnUnknown  func(event, dbtb []byte)
}

func (rr FrontendStrebmDecoder) RebdAll(r io.Rebder) error {
	dec := NewDecoder(r)

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
		} else if bytes.Equbl(event, []byte("mbtches")) {
			if rr.OnMbtches == nil {
				continue
			}
			vbr d []eventMbtchUnmbrshbller
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode mbtches pbylobd: %w", err)
			}
			m := mbke([]EventMbtch, 0, len(d))
			for _, e := rbnge d {
				m = bppend(m, e.EventMbtch)
			}
			rr.OnMbtches(m)
		} else if bytes.Equbl(event, []byte("filters")) {
			if rr.OnFilters == nil {
				continue
			}
			vbr d []*EventFilter
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode filters pbylobd: %w", err)
			}
			rr.OnFilters(d)
		} else if bytes.Equbl(event, []byte("blert")) {
			if rr.OnAlert == nil {
				continue
			}
			vbr d EventAlert
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode blert pbylobd: %w", err)
			}
			rr.OnAlert(&d)
		} else if bytes.Equbl(event, []byte("error")) {
			if rr.OnError == nil {
				continue
			}
			vbr d EventError
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

type eventMbtchUnmbrshbller struct {
	EventMbtch
}

func (r *eventMbtchUnmbrshbller) UnmbrshblJSON(b []byte) error {
	vbr typeU struct {
		Type MbtchType `json:"type"`
	}

	if err := json.Unmbrshbl(b, &typeU); err != nil {
		return err
	}

	switch typeU.Type {
	cbse ContentMbtchType:
		r.EventMbtch = &EventContentMbtch{}
	cbse PbthMbtchType:
		r.EventMbtch = &EventPbthMbtch{}
	cbse RepoMbtchType:
		r.EventMbtch = &EventRepoMbtch{}
	cbse SymbolMbtchType:
		r.EventMbtch = &EventSymbolMbtch{}
	cbse CommitMbtchType:
		r.EventMbtch = &EventCommitMbtch{}
	defbult:
		return errors.Errorf("unknown MbtchType %v", typeU.Type)
	}
	return json.Unmbrshbl(b, r.EventMbtch)
}
