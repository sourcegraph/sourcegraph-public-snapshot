// Package csp provides a middleware for setting
// Content-Security-Policy headers and receiving CSP violation
// reports.
package csp

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"reflect"
	"strings"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httptrace"
)

// Handler sets Content-Security-Policy{,-Report-Only} headers and
// logs CSP reports to stderr.
type Handler struct {
	cfg Config

	csp           string // Content-Security-Policy header
	cspReportOnly string // Content-Security-Policy-Report-Only header

	// ReportLog is the logger that CSP violation reports are sent
	// to. If nil, the default logger (from package log) is used.
	ReportLog *log.Logger
}

// Middleware is the middleware implementation of this CSP handler.
func (h *Handler) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if h.csp != "" {
			w.Header().Set("Content-Security-Policy", h.csp)
		}
		if h.cspReportOnly != "" {
			w.Header().Set("Content-Security-Policy-Report-Only", h.cspReportOnly)
		}

		if r.Method == "POST" {
			requestURI := r.URL.RequestURI()
			if h.cfg.Policy != nil && requestURI == h.cfg.Policy.ReportURI {
				httptrace.SetRouteName(r, "middleware.csp")
				h.logCSPReport("Content-Security-Policy", h.cfg.Policy, r.Body)
				return
			} else if h.cfg.PolicyReportOnly != nil && requestURI == h.cfg.PolicyReportOnly.ReportURI {
				httptrace.SetRouteName(r, "middleware.csp")
				h.logCSPReport("Content-Security-Policy-Report-Only", h.cfg.Policy, r.Body)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// compile precompiles the Content-Security-Policy{,-Report-Only}
// headers that h sets.
func (h *Handler) compile() {
	if h.cfg.Policy != nil {
		h.csp = h.cfg.Policy.compile()
	}
	if h.cfg.PolicyReportOnly != nil {
		h.cspReportOnly = h.cfg.PolicyReportOnly.compile()
	}
}

func (h *Handler) logCSPReport(name string, policy *Policy, body io.ReadCloser) {
	var logFn func(format string, v ...interface{})
	if h.ReportLog != nil {
		logFn = h.ReportLog.Printf
	} else {
		logFn = log.Printf
	}

	data, err := ioutil.ReadAll(io.LimitReader(body, 1024*1024 /* 1 MB */))
	if err != nil {
		logFn("CSP violation log: failed to read body: %s", err)
		return
	}
	defer body.Close()

	logFn("CSP violation report (%s): %s (for policy %+v)", name, data, policy)
}

// NewHandler creates a new handler that adds the configured Content
// Security Policy (CSP) headers to HTTP responses. If cfg is invalid,
// NewHandler returns an error.
func NewHandler(cfg Config) *Handler {
	h := &Handler{cfg: cfg}
	h.compile()
	return h
}

// Config configures the Content Security Policy headers sent by
// Handler.
type Config struct {
	Policy           *Policy // if non-nil, configures the Content-Security-Policy header
	PolicyReportOnly *Policy // if non-nil, configures the Content-Security-Policy-Report-Only header
}

// Policy describes a Content Security Policy.
type Policy struct {
	DefaultSrc []string
	ScriptSrc  []string
	ObjectSrc  []string
	StyleSrc   []string
	ImgSrc     []string
	MediaSrc   []string
	FrameSrc   []string
	FontSrc    []string
	ConnectSrc []string
	ReportURI  string
}

// compile compiles a Policy into its corresponding CSP header value.
func (p Policy) compile() string {
	var directives []string
	v, t := reflect.ValueOf(p), reflect.TypeOf(p)
	for i := 0; i < t.NumField(); i++ {
		fv := v.Field(i)
		ft := t.Field(i)
		if fv.Len() != 0 {
			if ft.Name == "ReportURI" {
				directives = append(directives, directive(ft.Name, []string{fv.Interface().(string)}))
			} else {
				directives = append(directives, directive(ft.Name, fv.Interface().([]string)))
			}
		}
	}
	return strings.Join(directives, "; ")
}

func directive(fieldName string, vals []string) string {
	var dirName string
	if fieldName == "ReportURI" {
		dirName = "report-uri"
	} else {
		dirName = strings.ToLower(strings.TrimSuffix(fieldName, "Src")) + "-src"
	}
	return dirName + " " + strings.Join(vals, " ")
}
