package http

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
)

func TestFrontendClient(t *testing.T) {
	type Event struct {
		Name  string
		Value any
	}

	want := []Event{{
		Name: "progress",
		Value: &api.Progress{
			MatchCount: 5,
		},
	}, {
		Name: "progress",
		Value: &api.Progress{
			MatchCount: 10,
		},
	}, {
		Name: "matches",
		Value: []EventMatch{
			&EventContentMatch{
				Type: ContentMatchType,
				Path: "test",
			},
			&EventPathMatch{
				Type: PathMatchType,
				Path: "test",
			},
			&EventRepoMatch{
				Type:       RepoMatchType,
				Repository: "test",
			},
			&EventSymbolMatch{
				Type: SymbolMatchType,
				Path: "test",
			},
			&EventCommitMatch{
				Type:   CommitMatchType,
				Detail: "test",
			},
		},
	}, {
		Name: "filters",
		Value: []*EventFilter{{
			Value: "filter-1",
		}, {
			Value: "filter-2",
		}},
	}, {
		Name: "alert",
		Value: &EventAlert{
			Title: "alert",
		},
	}, {
		Name: "error",
		Value: &EventError{
			Message: "error",
		},
	}}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ew, err := NewWriter(w)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		for _, e := range want {
			ew.Event(e.Name, e.Value)
		}
		ew.Event("done", struct{}{})
	}))
	defer ts.Close()

	req, err := NewRequest(ts.URL, "hello world")
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var got []Event
	err = FrontendStreamDecoder{
		OnProgress: func(d *api.Progress) {
			got = append(got, Event{Name: "progress", Value: d})
		},
		OnMatches: func(d []EventMatch) {
			got = append(got, Event{Name: "matches", Value: d})
		},
		OnFilters: func(d []*EventFilter) {
			got = append(got, Event{Name: "filters", Value: d})
		},
		OnAlert: func(d *EventAlert) {
			got = append(got, Event{Name: "alert", Value: d})
		},
		OnError: func(d *EventError) {
			got = append(got, Event{Name: "error", Value: d})
		},
		OnUnknown: func(event, data []byte) {
			t.Fatalf("got unexpected event: %s %s", event, data)
		},
	}.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	if d := cmp.Diff(want, got); d != "" {
		t.Fatalf("mismatch (-want +got):\n%s", d)
	}
}

func TestNewRequestWithVersion(t *testing.T) {
	baseURL := "http://example.com"
	patternTypeKeyword := query.SearchTypeKeyword

	tests := []struct {
		name          string
		query         string
		version       string
		patternType   *query.SearchType
		expectedQuery string
	}{
		{
			name:          "No version, no patternType",
			query:         "test",
			version:       "",
			patternType:   nil,
			expectedQuery: "q=test",
		},
		{
			name:          "Only version",
			query:         "test",
			version:       "V4",
			patternType:   nil,
			expectedQuery: "q=test&v=V4",
		},
		{
			name:          "Only patternType",
			query:         "test",
			version:       "",
			patternType:   &patternTypeKeyword,
			expectedQuery: "q=test&t=keyword",
		},
		{
			name:          "Version and patternType",
			query:         "test query",
			version:       "V3",
			patternType:   &patternTypeKeyword,
			expectedQuery: "q=test+query&v=V3&t=keyword",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := NewRequestWithVersion(baseURL, tt.query, tt.version, tt.patternType)
			require.NoError(t, err)

			// Check the request method
			require.Equal(t, "GET", req.Method)

			// Check the request URL
			parsedURL, err := url.Parse(req.URL.String())
			require.NoError(t, err)

			expectedBaseURL, err := url.Parse(baseURL)
			require.NoError(t, err)
			require.Equal(t, expectedBaseURL.Host, parsedURL.Host)
			require.Equal(t, expectedBaseURL.Scheme, parsedURL.Scheme)
			require.Equal(t, "/search/stream", parsedURL.Path)

			// Check the query parameters
			queryParams := parsedURL.Query()
			expectedParams, err := url.ParseQuery(tt.expectedQuery)
			require.NoError(t, err)

			require.Equal(t, expectedParams, queryParams)

			// Check the Accept header
			require.Equal(t, "text/event-stream", req.Header.Get("Accept"))
		})
	}
}
