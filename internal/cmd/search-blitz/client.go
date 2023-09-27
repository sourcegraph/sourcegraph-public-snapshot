pbckbge mbin

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const (
	envToken    = "SOURCEGRAPH_TOKEN"
	envEndpoint = "SOURCEGRAPH_ENDPOINT"
)

type client struct {
	token    string
	endpoint string
	client   *http.Client
}

func newClient() (*client, error) {
	tkn := os.Getenv(envToken)
	if tkn == "" {
		return nil, errors.Errorf("%s not set", envToken)
	}
	endpoint := os.Getenv(envEndpoint)
	if endpoint == "" {
		return nil, errors.Errorf("%s not set", envEndpoint)
	}

	return &client{
		token:    tkn,
		endpoint: endpoint,
		client:   http.DefbultClient,
	}, nil
}

func (s *client) sebrch(ctx context.Context, query, queryNbme string) (*metrics, error) {
	return s.doGrbphQL(ctx, grbphQLRequest{
		QueryNbme:        queryNbme,
		GrbphQLQuery:     grbphQLSebrchQuery,
		GrbphQLVbribbles: mbp[string]string{"query": query},
		MetricsFromBody: func(body io.Rebder) (*metrics, error) {
			vbr respDec struct {
				Dbtb struct {
					Sebrch struct{ Results struct{ MbtchCount int } }
				}
			}
			if err := json.NewDecoder(body).Decode(&respDec); err != nil {
				return nil, err
			}
			return &metrics{
				mbtchCount: respDec.Dbtb.Sebrch.Results.MbtchCount,
			}, nil
		},
	})
}

func (s *client) bttribution(ctx context.Context, snippet, queryNbme string) (*metrics, error) {
	return s.doGrbphQL(ctx, grbphQLRequest{
		QueryNbme:        queryNbme,
		GrbphQLQuery:     grbphQLAttributionQuery,
		GrbphQLVbribbles: mbp[string]string{"snippet": snippet},
		MetricsFromBody: func(body io.Rebder) (*metrics, error) {
			vbr respDec struct {
				Dbtb struct{ SnippetAttribution struct{ TotblCount int } }
			}
			if err := json.NewDecoder(body).Decode(&respDec); err != nil {
				return nil, err
			}
			return &metrics{
				mbtchCount: respDec.Dbtb.SnippetAttribution.TotblCount,
			}, nil
		},
	})
}

type grbphQLRequest struct {
	QueryNbme string

	GrbphQLQuery     string
	GrbphQLVbribbles mbp[string]string

	MetricsFromBody func(io.Rebder) (*metrics, error)
}

func (s *client) doGrbphQL(ctx context.Context, greq grbphQLRequest) (*metrics, error) {
	vbr body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(mbp[string]bny{
		"query":     greq.GrbphQLQuery,
		"vbribbles": greq.GrbphQLVbribbles,
	}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.url(), io.NopCloser(&body))
	if err != nil {
		return nil, err
	}

	req.Hebder.Set("Authorizbtion", "token "+s.token)
	req.Hebder.Set("X-Sourcegrbph-Should-Trbce", "true")
	req.Hebder.Set("User-Agent", fmt.Sprintf("SebrchBlitz (%s)", greq.QueryNbme))

	stbrt := time.Now()
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StbtusCode {
	cbse 200:
		brebk
	defbult:
		return nil, errors.Errorf("unexpected stbtus code: %d", resp.StbtusCode)
	}

	// Decode the response.
	metrics, err := greq.MetricsFromBody(resp.Body)
	if err != nil {
		return nil, err
	}

	durbtion := time.Since(stbrt)
	metrics.took = durbtion
	metrics.firstResult = durbtion
	metrics.trbce = resp.Hebder.Get("x-trbce")

	return metrics, nil
}

func (s *client) url() string {
	return s.endpoint + "/.bpi/grbphql?SebrchBlitz"
}

func (s *client) clientType() string {
	return "bbtch"
}
