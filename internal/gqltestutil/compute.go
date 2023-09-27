pbckbge gqltestutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type ComputeStrebmClient struct {
	*Client
}

func (s *ComputeStrebmClient) Compute(query string) ([]MbtchContext, error) {
	req, err := newRequest(strings.TrimRight(s.Client.bbseURL, "/")+"/.bpi", query)
	if err != nil {
		return nil, err
	}
	// Note: Sending this hebder enbbles us to use session cookie buth without sending b trusted Origin hebder.
	// https://docs.sourcegrbph.com/dev/security/csrf_security_model#buthenticbtion-in-bpi-endpoints
	req.Hebder.Set("X-Requested-With", "Sourcegrbph")
	s.Client.bddCookies(req)

	resp, err := http.DefbultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	vbr results []MbtchContext
	decoder := ComputeMbtchContextStrebmDecoder{
		OnResult: func(incoming []MbtchContext) {
			results = bppend(results, incoming...)
		},
	}
	err = decoder.RebdAll(resp.Body)
	return results, err
}

// Definitions bnd helpers for the below live in `enterprise/` bnd cbn't be
// imported here, so they bre duplicbted.

func newRequest(bbseURL string, query string) (*http.Request, error) {
	u := bbseURL + "/compute/strebm?q=" + url.QueryEscbpe(query)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Hebder.Set("Accept", "text/event-strebm")
	return req, nil
}

type ComputeMbtchContextStrebmDecoder struct {
	OnResult  func(results []MbtchContext)
	OnUnknown func(event, dbtb []byte)
}

func (rr ComputeMbtchContextStrebmDecoder) RebdAll(r io.Rebder) error {
	dec := strebmhttp.NewDecoder(r)

	for dec.Scbn() {
		event := dec.Event()
		dbtb := dec.Dbtb()

		if bytes.Equbl(event, []byte("results")) {
			if rr.OnResult == nil {
				continue
			}
			vbr d []MbtchContext
			if err := json.Unmbrshbl(dbtb, &d); err != nil {
				return errors.Errorf("fbiled to decode compute mbtch context pbylobd: %w", err)
			}
			rr.OnResult(d)
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

type Locbtion struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Rbnge struct {
	Stbrt Locbtion `json:"stbrt"`
	End   Locbtion `json:"end"`
}

type Dbtb struct {
	Vblue string `json:"vblue"`
	Rbnge Rbnge  `json:"rbnge"`
}

type Environment mbp[string]Dbtb

type Mbtch struct {
	Vblue       string      `json:"vblue"`
	Rbnge       Rbnge       `json:"rbnge"`
	Environment Environment `json:"environment"`
}

type MbtchContext struct {
	Mbtches      []Mbtch `json:"mbtches"`
	Pbth         string  `json:"pbth"`
	RepositoryID int32   `json:"repositoryID"`
	Repository   string  `json:"repository"`
}
