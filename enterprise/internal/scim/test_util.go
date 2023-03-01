package scim

import (
	"io"
	"net/http"
	"strings"
)

// createDummyRequest creates a dummy request with a body that is not empty.
func createDummyRequest() *http.Request {
	return &http.Request{Body: io.NopCloser(strings.NewReader("test"))}
}
