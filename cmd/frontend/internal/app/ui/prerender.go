package ui

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/jscontext"
	"github.com/sourcegraph/sourcegraph/internal/httpcli"
)

// TODO(sqs): use env var or whatever
const prerenderURL = "http://localhost:3190"

var prerenderHTTPClientDoer, _ = httpcli.NewInternalClientFactory("prerender").Doer()

type prerenderRequest struct {
	RequestURI string              `json:"requestURI"`
	JSContext  jscontext.JSContext `json:"jscontext"`
}

type prerenderResponse struct {
	HTML         string      `json:"html,omitempty"`
	InitialState interface{} `json:"initialState,omitempty"`
	RedirectURL  string      `json:"redirectURL,omitempty"`
	Error        string      `json:"error,omitempty"`
}

func prerender(ctx context.Context, req prerenderRequest) (_ *prerenderResponse, err error) {
	defer func() {
		if err != nil {
			stack := fmt.Sprintf("prerender url=%q", req.RequestURI)
			err = errors.Wrap(err, stack)
		}
	}()

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", prerenderURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	httpResp, err := prerenderHTTPClientDoer.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()
	if httpResp.StatusCode != http.StatusOK {
		return nil, errors.WithMessagef(err, "http status %d", httpResp.StatusCode)
	}

	var resp prerenderResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, errors.WithMessage(err, "decoding prerender response")
	}
	if resp.Error != "" {
		return nil, errors.New(resp.Error)
	}
	return &resp, nil
}
