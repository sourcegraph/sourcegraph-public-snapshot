package graphqlbackend

import (
	"net/url"
	"strings"
)

// pathEscapeExceptSlashes percent-encodes a URL path segment like
// url.PathEscape does except that it leaves slashes as-is.
func pathEscapeExceptSlashes(path string) string {
	return strings.Replace(url.PathEscape(path), "%2F", "/", -1)
}
