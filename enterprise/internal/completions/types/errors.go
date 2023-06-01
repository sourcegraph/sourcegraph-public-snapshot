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

// WriteResponseHeaders writes the error code and headers to the response.
// It does not write the response body, to allow different handlers to provide
// the message in different formats.
func (e *ErrStatusNotOK) WriteResponseHeaders(w http.ResponseWriter) {
	w.WriteHeader(e.statusCode)
	for k, vs := range e.responseHeader {
		for _, v := range vs {
			w.Header().Set(k, v)
		}
	}
}
