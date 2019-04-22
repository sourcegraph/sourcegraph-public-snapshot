package httpcli

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/kylelemons/godebug/pretty"
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
			name: "no context error, no doer error",
			cli:  newFakeClient(http.StatusOK, nil, nil),
			err:  "<nil>",
		},
		{
			name: "no context error, with doer error",
			cli:  newFakeClient(http.StatusOK, nil, errors.New("boom")),
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

func TestNewCertPool(t *testing.T) {
	pool := x509.NewCertPool()
	for _, tc := range []struct {
		name   string
		pool   *x509.CertPool
		cli    *http.Client
		assert func(testing.TB, *http.Client)
		err    string
	}{
		{
			name: "sets default transport if nil",
			cli:  &http.Client{},
			assert: func(t testing.TB, cli *http.Client) {
				if cli.Transport == nil {
					t.Fatal("transport wasn't set")
				}
			},
		},
		{
			name: "fails if transport isn't an http.Transport",
			cli:  &http.Client{Transport: bogusTransport{}},
			err:  "httpcli.NewCertPoolOpt: http.Client.Transport is not an *http.Transport",
		},
		{
			name: "sets TLSClientConfig if nil",
			cli:  &http.Client{Transport: &http.Transport{}},
			assert: func(t testing.TB, cli *http.Client) {
				if cli.Transport.(*http.Transport).TLSClientConfig == nil {
					t.Fatal("TLSClientConfig wasn't set")
				}
			},
		},
		{
			name: "pool is set to what is given",
			cli:  &http.Client{Transport: &http.Transport{}},
			pool: pool,
			assert: func(t testing.TB, cli *http.Client) {
				have := cli.Transport.(*http.Transport).TLSClientConfig.RootCAs
				if want := pool; !reflect.DeepEqual(have, want) {
					t.Fatal(pretty.Compare(have, want))
				}
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := NewCertPoolOpt(tc.pool)(tc.cli)

			if tc.err == "" {
				tc.err = "<nil>"
			}

			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("have error: %q\nwant error: %q", have, want)
			}

			if tc.assert != nil {
				tc.assert(t, tc.cli)
			}
		})
	}
}

func newFakeClient(code int, body []byte, err error) Doer {
	return DoerFunc(func(r *http.Request) (*http.Response, error) {
		rr := httptest.NewRecorder()
		_, _ = rr.Write(body)
		rr.WriteHeader(code)
		return rr.Result(), err
	})
}

type bogusTransport struct{}

func (t bogusTransport) RoundTrip(*http.Request) (*http.Response, error) {
	panic("should not be called")
}
