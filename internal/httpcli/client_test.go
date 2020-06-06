package httpcli

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestHeadersMiddleware(t *testing.T) {
	headers := []string{"X-Foo", "bar", "X-Bar", "foo"}
	for _, tc := range []struct {
		name    string
		cli     Doer
		headers []string
		err     string
	}{
		{
			name:    "odd number of headers panics",
			headers: headers[:1],
			cli: DoerFunc(func(r *http.Request) (*http.Response, error) {
				t.Fatal("should not be called")
				return nil, nil
			}),
			err: "missing header values",
		},
		{
			name:    "even number of headers are set",
			headers: headers,
			cli: DoerFunc(func(r *http.Request) (*http.Response, error) {
				for i := 0; i < len(headers); i += 2 {
					name := headers[i]
					if have, want := r.Header.Get(name), headers[i+1]; have != want {
						t.Errorf("header %q: have: %q, want: %q", name, have, want)
					}
				}
				return nil, nil
			}),
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.err == "" {
				tc.err = "<nil>"
			}

			defer func() {
				if err := recover(); err != nil {
					if have, want := fmt.Sprint(err), tc.err; have != want {
						t.Fatalf("have error: %q\nwant error: %q", have, want)
					}
				}
			}()

			cli := HeadersMiddleware(tc.headers...)(tc.cli)
			req, _ := http.NewRequest("GET", "http://dev/null", nil)

			_, err := cli.Do(req)
			if have, want := fmt.Sprint(err), tc.err; have != want {
				t.Fatalf("have error: %q\nwant error: %q", have, want)
			}
		})
	}
}

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

func genCert(subject string) (string, error) {
	priv, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{subject},
		},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	if err := pem.Encode(&b, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return "", err
	}
	return b.String(), nil
}

func TestNewCertPool(t *testing.T) {
	subject := "newcertpooltest"
	cert, err := genCert(subject)
	if err != nil {
		t.Fatal(err)
	}

	for _, tc := range []struct {
		name   string
		certs  []string
		cli    *http.Client
		assert func(testing.TB, *http.Client)
		err    string
	}{
		{
			name:  "fails if transport isn't an http.Transport",
			cli:   &http.Client{Transport: bogusTransport{}},
			certs: []string{cert},
			err:   "httpcli.NewCertPoolOpt: http.Client.Transport is not an *http.Transport: httpcli.bogusTransport",
		},
		{
			name:  "pool is set to what is given",
			cli:   &http.Client{Transport: &http.Transport{}},
			certs: []string{cert},
			assert: func(t testing.TB, cli *http.Client) {
				pool := cli.Transport.(*http.Transport).TLSClientConfig.RootCAs
				for _, have := range pool.Subjects() {
					if bytes.Contains(have, []byte(subject)) {
						return
					}
				}
				t.Fatal("could not find subject in pool")
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := NewCertPoolOpt(tc.certs...)(tc.cli)

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

func TestNewIdleConnTimeoutOpt(t *testing.T) {
	timeout := 33 * time.Second
	for _, tc := range []struct {
		name    string
		cli     *http.Client
		timeout time.Duration
		assert  func(testing.TB, *http.Client)
		err     string
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
			err:  "httpcli.NewIdleConnTimeoutOpt: http.Client.Transport is not an *http.Transport: httpcli.bogusTransport",
		},
		{
			name:    "IdleConnTimeout is set to what is given",
			cli:     &http.Client{Transport: &http.Transport{}},
			timeout: timeout,
			assert: func(t testing.TB, cli *http.Client) {
				have := cli.Transport.(*http.Transport).IdleConnTimeout
				if want := timeout; !reflect.DeepEqual(have, want) {
					t.Fatal(cmp.Diff(have, want))
				}
			},
		},
	} {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			err := NewIdleConnTimeoutOpt(tc.timeout)(tc.cli)

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

func TestNewTimeoutOpt(t *testing.T) {
	var cli http.Client

	timeout := 42 * time.Second
	err := NewTimeoutOpt(timeout)(&cli)
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}

	if have, want := cli.Timeout, timeout; have != want {
		t.Errorf("have Timeout %s, want %s", have, want)
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
