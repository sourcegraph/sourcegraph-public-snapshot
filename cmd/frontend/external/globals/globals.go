// Package globals exports symbols from frontend/globals. See the parent
// package godoc for more information.
package globals

import (
	"context"
	"net/url"

	"github.com/sourcegraph/sourcegraph/internal/tenant"
)

func ExternalURL(ctx context.Context) *url.URL {
	tnt := tenant.FromContext(ctx)
	return externalURLs[tnt.ID()]
}

var externalURLs = map[int]*url.URL{
	1: mustParseURL("https://erik.sourcegraph.test:3443"),
	2: mustParseURL("https://sourcegraph.test:3443"),
}

func mustParseURL(u string) *url.URL {
	url, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return url
}
