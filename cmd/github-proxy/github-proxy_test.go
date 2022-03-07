package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

func TestInstrumentHandler(_ *testing.T) {
	h := http.Handler(nil)
	instrumentHandler(prometheus.DefaultRegisterer, h)
}

func TestGitHubProxy(t *testing.T) {
	ch := make(chan struct{})
	blocking := make(chan struct{})
	p := &githubProxy{client: doerFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/block":
			select {
			case <-ch:
			default:
				close(blocking)
				<-ch
			}
		}

		return &http.Response{
			StatusCode: 200,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("body")),
		}, nil
	})}

	srv := httptest.NewServer(p)
	t.Cleanup(srv.Close)

	go func() {
		req, _ := http.NewRequest("GET", srv.URL+"/block", nil)
		req.Header.Add("Authorization", "user1")
		http.DefaultClient.Do(req) // blocks
	}()

	t.Run("locked", func(t *testing.T) {
		<-blocking

		req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)
		req.Header.Add("Authorization", "user1")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		_, err := http.DefaultClient.Do(req.WithContext(ctx))
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal(err)
		}
	})

	t.Run("different user", func(t *testing.T) {
		req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)

		// Different user request can go through
		req.Header.Set("Authorization", "Bearer user2")
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("want status code 200, got %v", resp.StatusCode)
		}
	})

	t.Run("unlocked", func(t *testing.T) {
		// Now the first user's request will finish, we can go through.
		close(ch)

		req, _ := http.NewRequest("GET", srv.URL+"/ok", nil)
		req.Header.Set("Authorization", "Bearer user1")

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		resp, err := http.DefaultClient.Do(req.WithContext(ctx))
		if err != nil {
			t.Fatal(err)
		}

		if resp.StatusCode != 200 {
			t.Fatalf("want status code 200, got %v", resp.StatusCode)
		}
	})
}

type doerFunc func(*http.Request) (*http.Response, error)

func (do doerFunc) Do(r *http.Request) (*http.Response, error) {
	return do(r)
}
