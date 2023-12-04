package client

import (
	"bytes"
	"encoding/json"
	"io"
	stdhttp "net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/compute"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// NewComputeStreamRequest returns an http.Request against the streaming API for query.
func NewComputeStreamRequest(baseURL string, query string) (*stdhttp.Request, error) {
	u := baseURL + "/compute/stream?q=" + url.QueryEscape(query)
	req, err := stdhttp.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

type ComputeMatchContextStreamDecoder struct {
	OnProgress func(*api.Progress)
	OnResult   func(results []compute.MatchContext)
	OnAlert    func(*http.EventAlert)
	OnError    func(*http.EventError)
	OnUnknown  func(event, data []byte)
}

func (rr ComputeMatchContextStreamDecoder) ReadAll(r io.Reader) error {
	dec := http.NewDecoder(r)

	for dec.Scan() {
		event := dec.Event()
		data := dec.Data()

		if bytes.Equal(event, []byte("progress")) {
			if rr.OnProgress == nil {
				continue
			}
			var d api.Progress
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode progress payload: %w", err)
			}
			rr.OnProgress(&d)
		} else if bytes.Equal(event, []byte("results")) {
			if rr.OnResult == nil {
				continue
			}
			var d []compute.MatchContext
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode compute match context payload: %w", err)
			}
			rr.OnResult(d)
		} else if bytes.Equal(event, []byte("alert")) {
			// This decoder can handle alerts, but at the moment the only alert that is returned by
			// the compute stream is if a query times out after 60 seconds.
			if rr.OnAlert == nil {
				continue
			}
			var d http.EventAlert
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode alert payload: %w", err)
			}
			rr.OnAlert(&d)
		} else if bytes.Equal(event, []byte("error")) {
			if rr.OnError == nil {
				continue
			}
			var d http.EventError
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode error payload: %w", err)
			}
			rr.OnError(&d)
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
