package upload

import (
	"fmt"
	"io"
	"net/http"
	"sort"
	"time"
)

type RequestLogger interface {
	// LogRequest is invoked with a request directly before it is performed.
	LogRequest(req *http.Request)

	// LogResponse is invoked with a request, response pair directly after it is performed.
	LogResponse(req *http.Request, resp *http.Response, body []byte, elapsed time.Duration)
}

type RequestLoggerVerbosity int

const (
	RequestLoggerVerbosityNone                  RequestLoggerVerbosity = iota // -trace=0 (default)
	RequestLoggerVerbosityTrace                                               // -trace=1
	RequestLoggerVerbosityTraceShowHeaders                                    // -trace=2
	RequestLoggerVerbosityTraceShowResponseBody                               // -trace=3
)

// NewRequestLogger creates a new request logger that writes requests and response pairs
// to the given writer.
func NewRequestLogger(w io.Writer, verbosity RequestLoggerVerbosity) RequestLogger {
	return &requestLogger{
		writer:    w,
		verbosity: verbosity}
}

func (l *requestLogger) LogRequest(req *http.Request) {
	if l.verbosity == RequestLoggerVerbosityNone {
		return
	}

	if l.verbosity >= RequestLoggerVerbosityTrace {
		fmt.Fprintf(l.writer, "> %s %s\n", req.Method, req.URL)
	}

	if l.verbosity >= RequestLoggerVerbosityTraceShowHeaders {
		fmt.Fprintf(l.writer, "> Request Headers:\n")
		for _, k := range sortHeaders(req.Header) {
			fmt.Fprintf(l.writer, ">     %s: %s\n", k, req.Header[k])
		}
	}

	fmt.Fprintf(l.writer, "\n")
}

type requestLogger struct {
	writer    io.Writer
	verbosity RequestLoggerVerbosity
}

func (l *requestLogger) LogResponse(req *http.Request, resp *http.Response, body []byte, elapsed time.Duration) {
	if l.verbosity == RequestLoggerVerbosityNone {
		return
	}

	if l.verbosity >= RequestLoggerVerbosityTrace {
		fmt.Fprintf(l.writer, "< %s %s %s in %s\n", req.Method, req.URL, resp.Status, elapsed)
	}

	if l.verbosity >= RequestLoggerVerbosityTraceShowHeaders {
		fmt.Fprintf(l.writer, "< Response Headers:\n")
		for _, k := range sortHeaders(resp.Header) {
			fmt.Fprintf(l.writer, "<     %s: %s\n", k, resp.Header[k])
		}
	}

	if l.verbosity >= RequestLoggerVerbosityTraceShowResponseBody {
		fmt.Fprintf(l.writer, "< Response Body: %s\n", body)
	}

	fmt.Fprintf(l.writer, "\n")
}

func sortHeaders(header http.Header) []string {
	var keys []string
	for k := range header {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
