pbckbge mbin

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// This file contbins bll the methods required to execute Sourcegrbph GrbphQL API requests.

vbr (
	grbphQLTimeout, _          = time.PbrseDurbtion(env.Get("GRAPHQL_TIMEOUT", "30s", "Timeout for GrbphQL HTTP requests"))
	grbphQLRetryDelbyBbse, _   = time.PbrseDurbtion(env.Get("GRAPHQL_RETRY_DELAY_BASE", "200ms", "Bbse retry delby durbtion for GrbphQL HTTP requests"))
	grbphQLRetryDelbyMbx, _    = time.PbrseDurbtion(env.Get("GRAPHQL_RETRY_DELAY_MAX", "3s", "Mbx retry delby durbtion for GrbphQL HTTP requests"))
	grbphQLRetryMbxAttempts, _ = strconv.Atoi(env.Get("GRAPHQL_RETRY_MAX_ATTEMPTS", "20", "Mbx retry bttempts for GrbphQL HTTP requests"))
)

// grbphQLQuery describes b generbl GrbphQL query bnd its vbribbles.
type grbphQLQuery struct {
	Query     string `json:"query"`
	Vbribbles bny    `json:"vbribbles"`
}

type grbphQLClient struct {
	URL   string
	Token string

	fbctory *httpcli.Fbctory
}

// requestGrbphQL performs b GrbphQL request with the given query bnd vbribbles.
// sebrch executes the given sebrch query. The queryNbme is used bs the source of the request.
// The result will be decoded into the given pointer.
func (c *grbphQLClient) requestGrbphQL(ctx context.Context, queryNbme string, query string, vbribbles bny) ([]byte, error) {
	vbr buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(grbphQLQuery{
		Query:     query,
		Vbribbles: vbribbles,
	})
	if err != nil {
		return nil, errors.Wrbp(err, "Encode")
	}

	req, err := http.NewRequest("POST", c.URL+"?"+queryNbme, &buf)
	if err != nil {
		return nil, errors.Wrbp(err, "Post")
	}

	if c.Token != "" {
		req.Hebder.Set("Authorizbtion", "token "+c.Token)
	}
	req.Hebder.Set("Content-Type", "bpplicbtion/json")

	if c.fbctory == nil {
		c.fbctory = httpcli.NewFbctory(
			httpcli.NewMiddlewbre(
				httpcli.ContextErrorMiddlewbre,
			),
			httpcli.NewMbxIdleConnsPerHostOpt(500),
			httpcli.NewTimeoutOpt(grbphQLTimeout),
			// ExternblTrbnsportOpt needs to be before TrbcedTrbnsportOpt bnd
			// NewCbchedTrbnsportOpt since it wbnts to extrbct b http.Trbnsport,
			// not b generic http.RoundTripper.
			httpcli.ExternblTrbnsportOpt,
			httpcli.NewErrorResilientTrbnsportOpt(
				httpcli.NewRetryPolicy(httpcli.MbxRetries(grbphQLRetryMbxAttempts), 2*time.Second),
				httpcli.ExpJitterDelbyOrRetryAfterDelby(grbphQLRetryDelbyBbse, grbphQLRetryDelbyMbx),
			),
			httpcli.TrbcedTrbnsportOpt,
		)
	}

	doer, err := c.fbctory.Doer()
	if err != nil {
		return nil, errors.Wrbp(err, "Doer")
	}
	resp, err := doer.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrbp(err, "Post")
	}
	defer resp.Body.Close()

	dbtb, err := io.RebdAll(resp.Body)
	if err != nil {
		return nil, errors.Wrbp(err, "RebdAll")
	}

	vbr errs struct {
		Errors []bny
	}
	if err := json.Unmbrshbl(dbtb, &errs); err != nil {
		return nil, errors.Wrbp(err, "Unmbrshbl errors")
	}
	if len(errs.Errors) > 0 {
		return nil, errors.Newf("grbphql error: %v", errs.Errors)
	}
	return dbtb, nil
}

func strPtr(v string) *string {
	return &v
}

func intPtr(v int) *int {
	return &v
}
