package httpapi

import (
	"net/http"

	"github.com/sourcegraph/sourcegraph/internal/featureflag"
)

func getStartedRedirect(w http.ResponseWriter, r *http.Request) error {
	flagset := featureflag.FromContext(r.Context())

	if flagset.GetBoolOr("ab_unified_registration", false) {
		http.Redirect(w, r, "https://sourcegraph.com/sign-up?returnTo=/deployment-options", 301)
	} else {
		http.Redirect(w, r, "https://about.sourcegraph.com/get-started", 301)
	}

	return nil
}
