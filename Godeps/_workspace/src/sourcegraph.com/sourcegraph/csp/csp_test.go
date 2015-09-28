package csp

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestHandler(t *testing.T) {
	tests := []struct {
		Config
		wantHeaders http.Header
	}{
		{
			Config{
				Policy:           &Policy{DefaultSrc: []string{"'self'", "foo.com"}},
				PolicyReportOnly: &Policy{ImgSrc: []string{"'self'", "https://bar.com"}, ReportURI: "/foo"},
			},
			http.Header{
				"Content-Security-Policy":             []string{"default-src 'self' foo.com"},
				"Content-Security-Policy-Report-Only": []string{"img-src 'self' https://bar.com; report-uri /foo"},
			},
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("GET", "/foo", nil)
		h := NewHandler(test.Config)
		h.ServeHTTP(w, r, nil)

		if !reflect.DeepEqual(w.HeaderMap, test.wantHeaders) {
			t.Errorf("Config %+v\ngot headers %v\nwant        %v", test.Config, w.HeaderMap, test.wantHeaders)
			continue
		}
	}
}

func TestHandlerReport(t *testing.T) {
	tests := []struct {
		Config
		wantReportPrefix string
	}{
		{
			Config{
				Policy: &Policy{ReportURI: "/csp-report"},
			},
			"CSP violation report (Content-Security-Policy):",
		},
	}
	for _, test := range tests {
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/csp-report", strings.NewReader("REPORT BODY"))
		h := NewHandler(test.Config)
		var buf bytes.Buffer
		h.ReportLog = log.New(&buf, "", 0)
		h.ServeHTTP(w, r, nil)

		if report := buf.String(); !strings.HasPrefix(report, test.wantReportPrefix) {
			t.Errorf("Config %+v: got report %q, want prefix %q", test.Config, report, test.wantReportPrefix)
		}
	}
}
