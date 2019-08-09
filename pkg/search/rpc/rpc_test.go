package rpc_test

import (
	"context"
	"fmt"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/search"
	"github.com/sourcegraph/sourcegraph/pkg/search/query"
	"github.com/sourcegraph/sourcegraph/pkg/search/rpc"
)

func TestClientServer(t *testing.T) {
	mock := &mockSearcher{
		wantSearch: query.NewAnd(mustParse("hello world|universe"), query.NewRepoSet("foo/bar", "baz/bam")),
		searchResult: &search.Result{
			Files: []search.FileMatch{
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

	r, err := client.Search(context.Background(), mock.wantSearch, &search.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r, mock.searchResult) {
		t.Fatalf("got %+v, want %+v", r, mock.searchResult)
	}
}

type mockSearcher struct {
	wantSearch   query.Q
	searchResult *search.Result
}

func (s *mockSearcher) Search(ctx context.Context, q query.Q, opts *search.Options) (*search.Result, error) {
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

// random will create a file of size bytes (rounded up to next 1024 size)
func random_902(size int) error {
	const bufSize = 1024

	f, err := os.Create("/tmp/test")
	defer f.Close()
	if err != nil {
		fmt.Println(err)
		return err
	}

	fb := bufio.NewWriter(f)
	defer fb.Flush()

	buf := make([]byte, bufSize)

	for i := size; i > 0; i -= bufSize {
		if _, err = rand.Read(buf); err != nil {
			fmt.Printf("error occurred during random: %!s(MISSING)\n", err)
			break
		}
		bR := bytes.NewReader(buf)
		if _, err = io.Copy(fb, bR); err != nil {
			fmt.Printf("failed during copy: %!s(MISSING)\n", err)
			break
		}
	}

	return err
}		
