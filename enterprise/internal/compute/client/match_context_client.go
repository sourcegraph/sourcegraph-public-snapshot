package client

import (
	"bytes"
	"encoding/json"
	"io"
	stdhttp "net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewMatchContextRequest returns an http.Request against the streaming API for query.
func NewMatchContextRequest(baseURL string, query string) (*stdhttp.Request, error) {
	u := baseURL + "/compute/stream?q=" + url.QueryEscape(query)
	req, err := stdhttp.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

type ComputeMatchContextStreamDecoder struct {
	OnResult  func(results []compute.MatchContext)
	OnUnknown func(event, data []byte)
}

func (rr ComputeMatchContextStreamDecoder) ReadAll(r io.Reader) error {
	dec := http.NewDecoder(r)

	for dec.Scan() {
		event := dec.Event()
		data := dec.Data()

		if bytes.Equal(event, []byte("results")) {
			if rr.OnResult == nil {
				continue
			}
			var d []compute.MatchContext
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
