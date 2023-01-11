package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

const maxPayloadSize = 10 * 1024 * 1024 // 10mb

// NewRequest returns an http.Request against the streaming API for query.
func NewRequest(baseURL string, query string) (*http.Request, error) {
	// when an empty string is passed as version, the route handler defaults to using the
	// latest supported version.
	return NewRequestWithVersion(baseURL, query, "")
}

// NewRequestWithVersion returns an http.Request against the streaming API for query with the specified version.
func NewRequestWithVersion(baseURL, query, version string) (*http.Request, error) {
	u := fmt.Sprintf("%s/search/stream?v=%s&q=%s", baseURL, version, url.QueryEscape(query))
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/event-stream")
	return req, nil
}

// FrontendStreamDecoder decodes streaming events from the frontend service
type FrontendStreamDecoder struct {
	OnProgress func(*api.Progress)
	OnMatches  func([]EventMatch)
	OnFilters  func([]*EventFilter)
	OnAlert    func(*EventAlert)
	OnError    func(*EventError)
	OnUnknown  func(event, data []byte)
}

func (rr FrontendStreamDecoder) ReadAll(r io.Reader) error {
	dec := NewDecoder(r)

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
		} else if bytes.Equal(event, []byte("matches")) {
			if rr.OnMatches == nil {
				continue
			}
			var d []eventMatchUnmarshaller
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode matches payload: %w", err)
			}
			m := make([]EventMatch, 0, len(d))
			for _, e := range d {
				m = append(m, e.EventMatch)
			}
			rr.OnMatches(m)
		} else if bytes.Equal(event, []byte("filters")) {
			if rr.OnFilters == nil {
				continue
			}
			var d []*EventFilter
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode filters payload: %w", err)
			}
			rr.OnFilters(d)
		} else if bytes.Equal(event, []byte("alert")) {
			if rr.OnAlert == nil {
				continue
			}
			var d EventAlert
			if err := json.Unmarshal(data, &d); err != nil {
				return errors.Errorf("failed to decode alert payload: %w", err)
			}
			rr.OnAlert(&d)
		} else if bytes.Equal(event, []byte("error")) {
			if rr.OnError == nil {
				continue
			}
			var d EventError
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

type eventMatchUnmarshaller struct {
	EventMatch
}

func (r *eventMatchUnmarshaller) UnmarshalJSON(b []byte) error {
	var typeU struct {
		Type MatchType `json:"type"`
	}

	if err := json.Unmarshal(b, &typeU); err != nil {
		return err
	}

	switch typeU.Type {
	case ContentMatchType:
		r.EventMatch = &EventContentMatch{}
	case PathMatchType:
		r.EventMatch = &EventPathMatch{}
	case RepoMatchType:
		r.EventMatch = &EventRepoMatch{}
	case SymbolMatchType:
		r.EventMatch = &EventSymbolMatch{}
	case CommitMatchType:
		r.EventMatch = &EventCommitMatch{}
	default:
		return errors.Errorf("unknown MatchType %v", typeU.Type)
	}
	return json.Unmarshal(b, r.EventMatch)
}
