package routevar

import "strings"

// pathUnescape is a limited version of url.QueryEscape that only unescapes '?'.
func pathUnescape(p string) string {
	return strings.ReplaceAll(p, "%3F", "?")
}
