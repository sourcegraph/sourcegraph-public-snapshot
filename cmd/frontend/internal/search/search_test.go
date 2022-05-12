package search

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	api2 "github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/run"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming/api"
	streamhttp "github.com/sourcegraph/sourcegraph/internal/search/streaming/http"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestServeStream_empty(t *testing.T) {
	graphqlbackend.MockDecodedViewerFinalSettings = &schema.Settings{}
	t.Cleanup(func() { graphqlbackend.MockDecodedViewerFinalSettings = nil })

	mock := client.NewMockSearchClient()
	mock.PlanFunc.SetDefaultReturn(&run.SearchInputs{}, nil)

	ts := httptest.NewServer(&streamHandler{
		flushTickerInternal: 1 * time.Millisecond,
		pingTickerInterval:  1 * time.Millisecond,
		searchClient:        mock,
	})
	defer ts.Close()

	res, err := http.Get(ts.URL + "?q=test")
	if err != nil {
		t.Fatal(err)
	}
	b, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		t.Fatal(err)
	}
	if res.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", res.StatusCode)
	}
	if testing.Verbose() {
		t.Logf("GET:\n%s", b)
	}
}

func TestDisplayLimit(t *testing.T) {
	cases := []struct {
		queryString         string
		displayLimit        int
		wantDisplayLimitHit bool
		wantMatchCount      int
		wantMessage         string
	}{
		{
			queryString:         "foo count:2",
			displayLimit:        1,
			wantDisplayLimitHit: true,
			wantMatchCount:      2,
			wantMessage:         "We only display 1 result even if your search returned more results. To see all results and configure the display limit, use our CLI.",
		},
		{
			queryString:         "foo count:2",
			displayLimit:        2,
			wantDisplayLimitHit: false,
			wantMatchCount:      2,
		},
		{
			queryString:         "foo count:2",
			displayLimit:        3,
			wantDisplayLimitHit: false,
			wantMatchCount:      2,
		},
		{
			queryString:         "foo count:100",
			displayLimit:        -1, // no display limit set by caller
			wantDisplayLimitHit: false,
			wantMatchCount:      2,
		},
		{
			queryString:         "foo count:1",
			displayLimit:        -1, // no display limit set by caller
			wantDisplayLimitHit: false,
			wantMatchCount:      1,
		},
	}

	// any returns item, true if skipped contains an item matching reason.
	any := func(reason api.SkippedReason, skipped []api.Skipped) (api.Skipped, bool) {
		for _, s := range skipped {
			if s.Reason == reason {
				return s, true
			}
		}
		return api.Skipped{}, false
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("q=%s;displayLimit=%d", c.queryString, c.displayLimit), func(t *testing.T) {
			graphqlbackend.MockDecodedViewerFinalSettings = &schema.Settings{}
			t.Cleanup(func() { graphqlbackend.MockDecodedViewerFinalSettings = nil })

			mockInput := make(chan streaming.SearchEvent)
			mock := client.NewMockSearchClient()
			mock.PlanFunc.SetDefaultHook(func(_ context.Context, _ string, _ *string, queryString string, _ search.Protocol, _ *schema.Settings, _ bool) (*run.SearchInputs, error) {
				q, err := query.Parse(queryString, query.SearchTypeLiteralDefault)
				require.NoError(t, err)
				return &run.SearchInputs{
					Query: q,
				}, nil
			})
			mock.ExecuteFunc.SetDefaultHook(func(_ context.Context, stream streaming.Sender, _ *run.SearchInputs) (*search.Alert, error) {
				event := <-mockInput
				stream.Send(event)
				return nil, nil
			})

			repos := database.NewStrictMockRepoStore()
			repos.MetadataFunc.SetDefaultHook(func(_ context.Context, ids ...api2.RepoID) (_ []*types.SearchedRepo, err error) {
				res := make([]*types.SearchedRepo, 0, len(ids))
				for _, id := range ids {
					res = append(res, &types.SearchedRepo{
						ID: id,
					})
				}
				return res, nil
			})
			db := database.NewStrictMockDB()
			db.ReposFunc.SetDefaultReturn(repos)

			ts := httptest.NewServer(&streamHandler{
				db:                  db,
				flushTickerInternal: 1 * time.Millisecond,
				pingTickerInterval:  1 * time.Millisecond,
				searchClient:        mock,
			})
			defer ts.Close()

			req, _ := streamhttp.NewRequest(ts.URL, c.queryString)
			if c.displayLimit != -1 {
				q := req.URL.Query()
				q.Add("display", strconv.Itoa(c.displayLimit))
				req.URL.RawQuery = q.Encode()
			}

			var displayLimitHit bool
			var message string
			var matchCount int
			decoder := streamhttp.FrontendStreamDecoder{
				OnProgress: func(progress *api.Progress) {
					if skipped, ok := any(api.DisplayLimit, progress.Skipped); ok {
						displayLimitHit = true
						message = skipped.Message
					}
					matchCount = progress.MatchCount
				},
			}

			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			// Consume events.
			g := errgroup.Group{}
			g.Go(func() error {
				return decoder.ReadAll(resp.Body)
			})

			// Send 2 repository matches.
			mockInput <- streaming.SearchEvent{
				Results: []result.Match{mkRepoMatch(1), mkRepoMatch(2)},
			}
			if err := g.Wait(); err != nil {
				t.Fatal(err)
			}

			if matchCount != c.wantMatchCount {
				t.Fatalf("got %d, want %d", matchCount, c.wantMatchCount)
			}

			if got := displayLimitHit; got != c.wantDisplayLimitHit {
				t.Fatalf("got %t, want %t", got, c.wantDisplayLimitHit)
			}

			if c.wantDisplayLimitHit {
				if got := message; got != c.wantMessage {
					t.Fatalf("got %s, want %s", got, c.wantMessage)
				}
			}
		})
	}
}

func mkRepoMatch(id int) *result.RepoMatch {
	return &result.RepoMatch{
		ID:   api2.RepoID(id),
		Name: api2.RepoName(fmt.Sprintf("repo%d", id)),
	}
}
