package httputil

import (
	"net/http"
	"strings"
)

// Kind of match to apply to the header check.
type headerMatchType int

const (
	hmEquals headerMatchType = iota
	hmStartsWith
	hmEndsWith
	hmContains
)

// Check if the specified header matches the test string, applying the header match type
// specified.
func headerMatch(hdr http.Header, nm string, matchType headerMatchType, test string) bool {
	// First get the header value
	val := hdr[http.CanonicalHeaderKey(nm)]
	if len(val) == 0 {
		return false
	}
	// Prepare the match test
	test = strings.ToLower(test)
	for _, v := range val {
		v = strings.Trim(strings.ToLower(v), " \n\t")
		switch matchType {
		case hmEquals:
			if v == test {
				return true
			}
		case hmStartsWith:
			if strings.HasPrefix(v, test) {
				return true
			}
		case hmEndsWith:
			if strings.HasSuffix(v, test) {
				return true
			}
		case hmContains:
			if strings.Contains(v, test) {
				return true
			}
		}
	}
	return false
}
