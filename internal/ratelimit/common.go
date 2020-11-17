package ratelimit

import (
	"net/url"
	"strings"
)

// normaliseURL will attempt to normalise rawURL.
// If there is an error parsing it, we'll just return rawURL lower cased.
func normaliseURL(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return strings.ToLower(rawURL)
	}
	parsed.Host = strings.ToLower(parsed.Host)
	if !strings.HasSuffix(parsed.Path, "/") {
		parsed.Path += "/"
	}
	return parsed.String()
}
