package ui

import (
	"io"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sourcegraph/sourcegraph/internal/featureflag"
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

var sveltekitEnabledRoutes = map[string]struct{}{
	routeSearch:           {},
	routeTree:             {},
	routeBlob:             {},
	routeRepo:             {},
	routeRepoSettings:     {},
	routeRepoCodeGraph:    {},
	routeRepoCommit:       {},
	routeRepoBranches:     {},
	routeRepoBatchChanges: {},
	routeRepoCommits:      {},
	routeRepoTags:         {},
	routeRepoCompare:      {},
	routeRepoStats:        {},
	routeRepoOwn:          {},
}

// useSvelteKit returns true if the route is configured to be supported by useSvelteKit
// (see svelteKitEnabledRoutes) and the 'enable-sveltekit' feature flag is set
func useSvelteKit(r *http.Request) bool {
	route := mux.CurrentRoute(r)
	if route == nil {
		return false
	}

	if _, ok := sveltekitEnabledRoutes[route.GetName()]; !ok {
		return false
	}

	return featureflag.FromContext(r.Context()).GetBoolOr("enable-sveltekit", false)
}

// renderSvelteKit writes SvelteKit's fallback page to the provided writer
func renderSvelteKit(dst io.Writer) error {
	file, err := assets.Provider.Assets().Open("_sk/index.html")
	if err != nil {
		return err
	}
	defer file.Close()
	io.Copy(dst, file)
	return nil
}
