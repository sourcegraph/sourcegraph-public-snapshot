package rpc_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/query"
	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/rpc"
)

func TestClientServer(t *testing.T) {
	mock := &mockSearcher{
		wantSearch: query.NewAnd(mustParse("hello world|universe"), query.NewRepoSet("foo/bar", "baz/bam")),
		searchResult: &zoekt.Result{
			Files: []zoekt.FileMatch{
				{Path: "bin.go"},
			},
		},
	}
	server, err := rpc.Server(mock)
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(server)
	defer ts.Close()

	u, err := url.Parse(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	client := rpc.Client(u.Host)
	defer client.Close()

	r, err := client.Search(context.Background(), mock.wantSearch, &zoekt.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r, mock.searchResult) {
		t.Fatalf("got %+v, want %+v", r, mock.searchResult)
	}
}

type mockSearcher struct {
	wantSearch   query.Q
	searchResult *zoekt.Result
}

func (s *mockSearcher) Search(ctx context.Context, q query.Q, opts *zoekt.Options) (*zoekt.Result, error) {
	if q.String() != s.wantSearch.String() {
		return nil, fmt.Errorf("got query %s != %s", q.String(), s.wantSearch.String())
	}
	return s.searchResult, nil
}

func (*mockSearcher) Close() {}

func (*mockSearcher) String() string {
	return "mockSearcher"
}

func mustParse(s string) query.Q {
	q, err := query.Parse(s)
	if err != nil {
		panic(err)
	}
	return q
}
