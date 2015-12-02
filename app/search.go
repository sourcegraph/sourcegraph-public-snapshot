package app

import (
	"net/url"

	"golang.org/x/net/context"

	"src.sourcegraph.com/sourcegraph/app/appconf"
)

func showSearchForm(ctx context.Context, query url.Values) bool {
	if _, ok := query["EnableSearch"]; ok {
		return true
	}

	if appconf.Flags.DisableSearch {
		return false
	}

	if appconf.Flags.CustomNavLayout != "" {
		return false
	}

	return false
}
