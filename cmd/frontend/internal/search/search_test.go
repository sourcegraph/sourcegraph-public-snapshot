package search

import (
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestServeStream_empty(t *testing.T) {
	mock := &mockSearchResolver{
		done: make(chan struct{}),
	}
	mock.Close()

	ts := httptest.NewServer(&streamHandler{
		newSearchResolver: func(context.Context, *graphqlbackend.SearchArgs) (searchResolver, error) {
			return mock, nil
		}})
	defer ts.Close()

	res, err := http.Get(ts.URL + "?q=test")
	if err != nil {
		t.Fatal(err)
	}
	b, err := ioutil.ReadAll(res.Body)
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

// Ensures graphqlbackend matches the interface we expect
func TestDefaultNewSearchResolver(t *testing.T) {
	_, err := defaultNewSearchResolver(context.Background(), &graphqlbackend.SearchArgs{
		Version:  "V2",
		Settings: &schema.Settings{},
	})
	if err != nil {
		t.Fatal(err)
	}
}

type mockSearchResolver struct {
	done chan struct{}
	c    chan<- []graphqlbackend.SearchResultResolver
}

func (h *mockSearchResolver) Results(ctx context.Context) (*graphqlbackend.SearchResultsResolver, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-h.done:
		return &graphqlbackend.SearchResultsResolver{
			UserSettings: &schema.Settings{},
		}, nil
	}
}
func (h *mockSearchResolver) SetResultChannel(c chan<- []graphqlbackend.SearchResultResolver) {
	h.c = c
}

func (h *mockSearchResolver) Send(r []graphqlbackend.SearchResultResolver) {
	h.c <- r
}

func (h *mockSearchResolver) Close() {
	close(h.done)
}
