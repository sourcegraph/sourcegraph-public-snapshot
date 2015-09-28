package app

import (
	"net/url"

	"src.sourcegraph.com/sourcegraph/app/internal/returnto"
	"src.sourcegraph.com/sourcegraph/app/router"
)

func urlToWithReturnTo(routeName, returnTo string) *url.URL {
	url := router.Rel.URLTo(routeName)
	returnto.SetOnURL(url, returnTo)
	return url
}
