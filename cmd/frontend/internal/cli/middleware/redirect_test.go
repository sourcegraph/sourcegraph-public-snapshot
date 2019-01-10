package middleware

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/conf"
	"github.com/sourcegraph/sourcegraph/schema"
)

func TestCanonicalURL(t *testing.T) {
	handle := func(t *testing.T, req *http.Request) (redirect string) {
		t.Helper()

		// In most real requests, only the URL's Path and RawQuery are not set. (See
		// (*http.Request).URL docs.)
		req.URL = &url.URL{Path: req.URL.Path, RawQuery: req.URL.RawQuery}

		h := CanonicalURL(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code >= 300 && rr.Code <= 399 {
			return rr.Header().Get("Location")
		}
		if want := http.StatusOK; rr.Code != want {
			t.Errorf("got response code %d, want %d", rr.Code, want)
		}
		return ""
	}

	tests := []struct {
		externalURL          string
		httpToHttpsRedirect  string
		canonicalURLRedirect string

		url             string
		xForwardedProto string

		wantRedirect string
	}{
		{
			externalURL:         "http://example.com",
			httpToHttpsRedirect: "off",
			url:                 "http://example.com/foo",
			wantRedirect:        "",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "off",
			url:                 "http://example.com/foo",
			wantRedirect:        "",
		},
		{
			externalURL:          "https://example.com",
			httpToHttpsRedirect:  "off",
			canonicalURLRedirect: "enabled",
			url:                  "http://other.example.com/foo",
			wantRedirect:         "https://example.com/foo",
		},
		{
			externalURL:          "http://example.com",
			httpToHttpsRedirect:  "off",
			canonicalURLRedirect: "enabled",
			url:                  "https://other.example.com/foo",
			wantRedirect:         "http://example.com/foo",
		},
		{
			externalURL:          "http://example.com",
			httpToHttpsRedirect:  "off",
			canonicalURLRedirect: "enabled",
			url:                  "http://example.com",
			wantRedirect:         "",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "off",
			url:                 "https://example.com/foo",
			wantRedirect:        "",
		},

		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "on",
			url:                 "http://example.com/foo",
			wantRedirect:        "https://example.com/foo",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "on",
			url:                 "http://other.example.com/foo",
			wantRedirect:        "https://example.com/foo",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "on",
			url:                 "http://example.com/foo",
			xForwardedProto:     "https", // not trusted
			wantRedirect:        "https://example.com/foo",
		},
		{
			externalURL:          "https://example.com",
			httpToHttpsRedirect:  "on",
			canonicalURLRedirect: "enabled",
			url:                  "http://other.example.com/foo",
			wantRedirect:         "https://example.com/foo",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "on",
			url:                 "https://example.com/foo",
			wantRedirect:        "", // no infinite redirect loop
		},

		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "load-balanced",
			url:                 "http://example.com/foo",
			xForwardedProto:     "http",
			wantRedirect:        "https://example.com/foo",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "load-balanced",
			url:                 "http://example.com/foo",
			xForwardedProto:     "https",
			wantRedirect:        "",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "load-balanced",
			url:                 "https://example.com/foo",
			xForwardedProto:     "http",
			wantRedirect:        "https://example.com/foo",
		},
		{
			externalURL:         "https://example.com",
			httpToHttpsRedirect: "load-balanced",
			url:                 "https://example.com/foo",
			xForwardedProto:     "https",
			wantRedirect:        "",
		},

		{
			externalURL:          "https://example.com",
			httpToHttpsRedirect:  "load-balanced",
			canonicalURLRedirect: "enabled",
			url:                  "http://example.com/foo",
			xForwardedProto:      "http",
			wantRedirect:         "https://example.com/foo",
		},
		{
			externalURL:          "https://example.com",
			httpToHttpsRedirect:  "load-balanced",
			canonicalURLRedirect: "enabled",
			url:                  "http://example.com/foo",
			xForwardedProto:      "https",
			wantRedirect:         "",
		},
		{
			externalURL:          "https://example.com",
			httpToHttpsRedirect:  "load-balanced",
			canonicalURLRedirect: "enabled",
			url:                  "http://other.example.com/foo",
			xForwardedProto:      "https",
			wantRedirect:         "https://example.com/foo",
		},
		{
			externalURL:          "https://example.com",
			httpToHttpsRedirect:  "load-balanced",
			canonicalURLRedirect: "enabled",
			url:                  "https://example.com/foo",
			xForwardedProto:      "http",
			wantRedirect:         "https://example.com/foo",
		},
		{
			externalURL:          "https://example.com",
			httpToHttpsRedirect:  "load-balanced",
			canonicalURLRedirect: "enabled",
			url:                  "https://example.com/foo",
			xForwardedProto:      "https",
			wantRedirect:         "",
		},
	}
	for i, test := range tests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			mock := &conf.Unified{Critical: schema.CriticalConfiguration{ExternalURL: test.externalURL, HttpToHttpsRedirect: test.httpToHttpsRedirect}}
			if test.canonicalURLRedirect != "" {
				mock.ExperimentalFeatures = &schema.ExperimentalFeatures{CanonicalURLRedirect: test.canonicalURLRedirect}
			}
			conf.Mock(mock)
			defer conf.Mock(nil)
			req := httptest.NewRequest("GET", test.url, nil)
			req.Header.Set("X-Forwarded-Proto", test.xForwardedProto)
			if redirect := handle(t, req); redirect != test.wantRedirect {
				t.Errorf("got redirect %q, want redirect %q", redirect, test.wantRedirect)
			}
		})
	}

	t.Run("httpToHttpsRedirect invalid value", func(t *testing.T) {
		conf.Mock(&conf.Unified{Critical: schema.CriticalConfiguration{HttpToHttpsRedirect: "invalid"}})
		defer conf.Mock(nil)
		h := CanonicalURL(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if want := http.StatusInternalServerError; rr.Code != want {
			t.Errorf("got response code %d, want %d", rr.Code, want)
		}
		if got, want := rr.Body.String(), "Misconfigured httpToHttpsRedirect"; !strings.Contains(got, want) {
			t.Errorf("got %q, want contains %q", got, want)
		}
	})

	t.Run("externalURL invalid value", func(t *testing.T) {
		conf.Mock(&conf.Unified{Critical: schema.CriticalConfiguration{ExternalURL: "invalid"}})
		defer conf.Mock(nil)
		h := CanonicalURL(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if want := http.StatusInternalServerError; rr.Code != want {
			t.Errorf("got response code %d, want %d", rr.Code, want)
		}
		if got, want := rr.Body.String(), "Misconfigured externalURL"; !strings.Contains(got, want) {
			t.Errorf("got %q, want contains %q", got, want)
		}
	})

	t.Run("experimentalFeatures.canonicalURLRedirect invalid value", func(t *testing.T) {
		conf.Mock(&conf.Unified{
			SiteConfiguration: schema.SiteConfiguration{ExperimentalFeatures: &schema.ExperimentalFeatures{CanonicalURLRedirect: "invalid"}},
			Critical:          schema.CriticalConfiguration{ExternalURL: "http://example.com"},
		})
		defer conf.Mock(nil)
		h := CanonicalURL(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
		req := httptest.NewRequest("GET", "/", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if want := http.StatusInternalServerError; rr.Code != want {
			t.Errorf("got response code %d, want %d", rr.Code, want)
		}
		if got, want := rr.Body.String(), "Misconfigured experimentalFeatures.canonicalURLRedirect"; !strings.Contains(got, want) {
			t.Errorf("got %q, want contains %q", got, want)
		}
	})
}

func TestParseStringOrBool(t *testing.T) {
	defaultValue := "default"
	// parsedValue -> stringOrBool
	cases := map[string]interface{}{
		defaultValue: nil,
		"":           "",
		"hi":         "hi",
		"on":         true,
		"off":        false,
	}
	for want, v := range cases {
		got := parseStringOrBool(v, defaultValue)
		if got != want {
			t.Errorf("parseStringOrBool(%q) got %q want %q", v, got, want)
		}
	}
}
