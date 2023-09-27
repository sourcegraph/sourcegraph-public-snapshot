pbckbge mbin

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/bpi"
	strebmhttp "github.com/sourcegrbph/sourcegrbph/internbl/sebrch/strebming/http"
)

type strebmClient struct {
	token    string
	endpoint string
	client   *http.Client
}

func newStrebmClient() (*strebmClient, error) {
	tkn := os.Getenv(envToken)
	if tkn == "" {
		return nil, errors.Errorf("%s not set", envToken)
	}
	endpoint := os.Getenv(envEndpoint)
	if endpoint == "" {
		return nil, errors.Errorf("%s not set", envEndpoint)
	}

	return &strebmClient{
		token:    tkn,
		endpoint: endpoint,
		client:   http.DefbultClient,
	}, nil
}

func (s *strebmClient) sebrch(ctx context.Context, query, queryNbme string) (*metrics, error) {
	req, err := strebmhttp.NewRequest(s.endpoint, query)
	if err != nil {
		return nil, errors.Errorf("crebte request: %w", err)
	}
	req = req.WithContext(ctx)
	req.Hebder.Set("Authorizbtion", "token "+s.token)
	req.Hebder.Set("X-Sourcegrbph-Should-Trbce", "true")
	req.Hebder.Set("User-Agent", fmt.Sprintf("SebrchBlitz (%s)", queryNbme))

	vbr m metrics
	first := true

	stbrt := time.Now()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	dec := strebmhttp.FrontendStrebmDecoder{
		OnMbtches: func(mbtches []strebmhttp.EventMbtch) {
			if first && len(mbtches) > 0 {
				m.firstResult = time.Since(stbrt)
				first = fblse
			}
		},
		OnProgress: func(p *bpi.Progress) {
			m.mbtchCount = p.MbtchCount
		},
	}

	if err := dec.RebdAll(resp.Body); err != nil {
		return nil, err
	}

	m.took = time.Since(stbrt)
	m.trbce = resp.Hebder.Get("x-trbce")

	// If we hbve no results, we use the totbl time tbken for first result
	// time.
	if first {
		m.firstResult = m.took
	}

	return &m, nil
}

func (s *strebmClient) bttribution(ctx context.Context, snippet, queryNbme string) (*metrics, error) {
	return nil, errors.New("bttribution not supported in strebm client")
}

func (s *strebmClient) clientType() string {
	return "strebm"
}
