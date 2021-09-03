package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/envvar"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/globals"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

var allowRobotsVar = env.Get("ROBOTS_TXT_ALLOW", "false", "allow search engines to index the site")

func robotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	allowRobots, _ := strconv.ParseBool(allowRobotsVar)
	robotsTxtHelper(w, allowRobots)
}

func robotsTxtHelper(w io.Writer, allowRobots bool) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "User-agent: *")
	if allowRobots {
		fmt.Fprintln(&buf, "Allow: /")
		if envvar.SourcegraphDotComMode() {
			fmt.Fprintln(&buf, "Sitemap: https://sourcegraph.com/sitemap.xml.gz")
		}
	} else {
		fmt.Fprintln(&buf, "Disallow: /")
	}
	fmt.Fprintln(&buf)
	_, _ = buf.WriteTo(w)
}

func sitemapXmlGz(w http.ResponseWriter, r *http.Request) {
	if envvar.SourcegraphDotComMode() {
		http.Redirect(w, r, "https://storage.googleapis.com/sitemap-sourcegraph-com/sitemap.xml.gz", http.StatusFound)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

func favicon(w http.ResponseWriter, r *http.Request) {
	url := assetsutil.URL("/img/favicon.png")

	// Add query parameter for cache busting.
	query := url.Query()
	query.Set("v", "2")
	url.RawQuery = query.Encode()
	path := url.String()

	if branding := globals.Branding(); branding.Favicon != "" {
		path = branding.Favicon
	}
	http.Redirect(w, r, path, http.StatusMovedPermanently)
}
