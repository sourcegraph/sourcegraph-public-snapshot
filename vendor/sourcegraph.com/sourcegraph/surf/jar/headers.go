package jar

import "net/http"

// NewMemoryHeaders creates and readers a new http.Header slice.
func NewMemoryHeaders() http.Header {
	return make(http.Header, 10)
}
