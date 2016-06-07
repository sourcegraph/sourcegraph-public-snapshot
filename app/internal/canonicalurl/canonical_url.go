// Package canonicalurl creates canonical URLs from request URLs by
// stripping extraneous query params, etc.
package canonicalurl

import (
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
)

// nonCanonicalQueryParams are query parameters that do not affect
// what's displayed on our site.
var nonCanonicalQueryParams = []string{
	"utm_source", "utm_medium", "utm_campaign", returnto.ParamName,
}

// FromURL returns the canonical URL for the given URL. If it is the
// same as the currentURL, the currentURL is returned unchanged.
func FromURL(currentURL *url.URL) *url.URL {
	canonicalQuery := currentURL.Query()
	for _, k := range nonCanonicalQueryParams {
		canonicalQuery.Del(k)
	}
	canonicalURL := *currentURL
	canonicalURL.RawQuery = canonicalQuery.Encode()
	return &canonicalURL
}

// FromQuery removes non-canonical parameters from v.
func FromQuery(v url.Values) {
	for _, k := range nonCanonicalQueryParams {
		v.Del(k)
	}
}
