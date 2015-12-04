package platform

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"src.sourcegraph.com/sourcegraph/platform/pctx"

	"golang.org/x/net/context"
)

// SetPlatformRequestURL translates a url from a particular platform app to
// the relative version of the URL expected by the registered platform router.
// From the perspective of the platform router, all URLs are relative paths
// where the root is "/".
func SetPlatformRequestURL(ctx context.Context, w http.ResponseWriter, r *http.Request, rCopy *http.Request) error {
	stripPrefix := pctx.BaseURI(ctx)
	u, err := url.Parse(stripPrefix)
	if err != nil {
		return err
	}
	stripPrefix = u.Path

	// The canonical URL for app root page does not have a trailing slash, so redirect.
	if rCopy.URL.Path == stripPrefix+"/" {
		baseURL := stripPrefix
		if rCopy.URL.RawQuery != "" {
			baseURL += rCopy.URL.RawQuery
		}
		http.Redirect(w, r, baseURL, http.StatusMovedPermanently)
		return nil
	}

	if p := strings.TrimPrefix(rCopy.URL.Path, stripPrefix); len(p) < len(r.URL.Path) {
		rCopy.URL.Path = p
		if rCopy.URL.Path == "" {
			rCopy.URL.Path = "/"
		}
	} else {
		return fmt.Errorf("could not load platform URL: URL path %q did not have prefix %q", rCopy.URL.Path, stripPrefix)
	}

	return nil
}
