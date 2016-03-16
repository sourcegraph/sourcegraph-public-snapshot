package app

import (
	"net/url"

	"sourcegraph.com/sourcegraph/sourcegraph/app/internal/returnto"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
)

func urlToWithReturnTo(routeName, returnTo string) *url.URL {
	url := router.Rel.URLTo(routeName)
	returnto.SetOnURL(url, returnTo)
	return url
}
