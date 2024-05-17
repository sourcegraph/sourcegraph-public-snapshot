package gqltestutil

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"

	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type ComputeStreamClient struct {
	*Client
}

func (s *ComputeStreamClient) Compute(query string) ([]MatchContext, error) {
	req, err := newRequest(strings.TrimRight(s.Client.baseURL, "/")+"/.api", query)
	if err != nil {
		return nil, err
	}
	// Note: Sending this header enables us to use session cookie auth without sending a trusted Origin header.
	// https://docs-legacy.sourcegraph.com/dev/security/csrf_security_model#authentication-in-api-endpoints
	req.Header.Set("X-Requested-With", "Sourcegraph")
	s.Client.addCookies(req)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []MatchContext
	decoder := ComputeMatchContextStreamDecoder{
		OnResult: func(incoming []MatchContext) {
			results = append(results, incoming...)
		},
	}
	err = decoder.ReadAll(resp.Body)
	return results, err
}

// Definitions and helpers for the below live in `enterprise/` and can't be
// imported here, so they are duplicated.

func newRequest(baseURL string, query string) (*http.Request, error) {
	u := baseURL + "/compute/stream?q=" + url.QueryEscape(query)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

type ComputeMatchContextStreamDecoder struct {
	OnResult  func(results []MatchContext)
	OnUnknown func(event, data []byte)
}

func (rr ComputeMatchContextStreamDecoder) ReadAll(r io.Reader) error {
	dec := streamhttp.NewDecoder(r)

	for dec.Scan() {
		event := dec.Event()
		data := dec.Data()

		if bytes.Equal(event, []byte("results")) {
			if rr.OnResult == nil {
				continue
			}
			var d []MatchContext
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode compute match context payload: %w", err)
			}
			rr.OnResult(d)
		} else if bytes.Equal(event, []byte("done")) {
			// Always the last event
			break
		} else {
			if rr.OnUnknown == nil {
				continue
			}
			rr.OnUnknown(event, data)
		}
	}
	return dec.Err()
}

type Location struct {
	Offset int `json:"offset"`
	Line   int `json:"line"`
	Column int `json:"column"`
}

type Range struct {
	Start Location `json:"start"`
	End   Location `json:"end"`
}

type Data struct {
	Value string `json:"value"`
	Range Range  `json:"range"`
}

type Environment map[string]Data

type Match struct {
	Value       string      `json:"value"`
	Range       Range       `json:"range"`
	Environment Environment `json:"environment"`
}

type MatchContext struct {
	Matches      []Match `json:"matches"`
	Path         string  `json:"path"`
	RepositoryID int32   `json:"repositoryID"`
	Repository   string  `json:"repository"`
}
