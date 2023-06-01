package types

import (
	"fmt"
	"io"
	"net/http"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// ErrStatusNotOK is returned when the server responds with a non-200 status code.
//
// Implementations of CompletionsClient should return this error with
// NewErrStatusNotOK the server responds with a non-OK status.
//
// Callers of CompletionsClient should check for this error with AsErrStatusNotOK
// and handle it appropriately, typically with (*ErrStatusNotOK).WriteResponse.
type ErrStatusNotOK struct {
	source     string
	statusCode int
	// responseBody is a truncated copy of the response body, read on a best-effort basis.
	responseBody   string
	responseHeader http.Header
}

var _ error = &ErrStatusNotOK{}

func (e *ErrStatusNotOK) Error() string {
	return fmt.Sprintf("%s: unexpected status code %d: %s",
		e.source, e.statusCode, e.responseBody)
}

// NewErrStatusNotOK parses reads resp body and closes it to return an ErrStatusNotOK
// based on the response.
func NewErrStatusNotOK(source string, resp *http.Response) error {
	// Callers shouldn't be using this function if the response is OK, but let's
	// sanity-check anyway.
	if resp.StatusCode == http.StatusOK {
		return nil
	}

	// Do a partial read of what we've got.
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))

	return &ErrStatusNotOK{
		source:         source,
		statusCode:     resp.StatusCode,
		responseBody:   string(respBody),
		responseHeader: resp.Header,
	}
}

func IsErrStatusNotOK(err error) (*ErrStatusNotOK, bool) {
	if err == nil {
		return nil, false
	}

	e := &ErrStatusNotOK{}
	if errors.As(err, &e) {
		return e, true
	}

	return nil, false
}

// WriteHeader writes the resolved error code and headers to the response.
// Currently, only certain allow-listed status codes are written back as-is -
// all other codes are written back as 503 to indicate the upstream service is
// available.
//
// It does not write the response body, to allow different handlers to provide
// the message in different formats.
func (e *ErrStatusNotOK) WriteHeader(w http.ResponseWriter) {
	for k, vs := range e.responseHeader {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}

	// WriteHeader must come last, since it flushes the headers.
	switch e.statusCode {
	// Only write back certain allow-listed status codes as-is - all other status
	// codes are written back as 503 to avoid potential confusions with Sourcegraph
	// status codes while indicating that the upstream service is unavailable.
	//
	// Currently, we only write back status code 429 as-is to help support
	// rate limit handling in clients.
	case http.StatusTooManyRequests:
		w.WriteHeader(e.statusCode)
	default:
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}
