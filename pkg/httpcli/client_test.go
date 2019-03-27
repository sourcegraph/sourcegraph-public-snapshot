package httpcli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContextErrorMiddleware(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	for _, tc := range []struct {
		name string
		cli  Doer
		ctx  context.Context
		err  string
	}{
		{
			name: "no context, with doer error",
			cli:  newFakeClient(http.StatusOK, nil, errors.New("boom")),
			err:  "boom",
		},
		{
			name: "with context and doer error",
			cli:  newFakeClient(http.StatusOK, nil, errors.New("boom")),
			ctx:  context.Background(),
			err:  "boom",
		},
		{
			name: "with context error and no doer error",
			cli:  newFakeClient(http.StatusOK, nil, nil),
			ctx:  cancelled,
			err:  "<nil>",
		},
		{
			name: "with context error and doer error",
			cli:  newFakeClient(http.StatusOK, nil, errors.New("boom")),
			ctx:  cancelled,
			err:  context.Canceled.Error(),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cli := ContextErrorMiddleware(tc.cli)

			req, _ := http.NewRequest("GET", "http://dev/null", nil)

			if tc.ctx != nil {
				req = req.WithContext(tc.ctx)
			}

			_, err := cli.Do(req)

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("have error: %q\nwant error: %q", have, want)
			}
		})
	}
}

func newFakeClient(code int, body []byte, err error) Doer {
	return DoerFunc(func(r *http.Request) (*http.Response, error) {
		rr := httptest.NewRecorder()
		rr.Write(body)
		rr.WriteHeader(code)
		return rr.Result(), err
	})
}
