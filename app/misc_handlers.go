package app

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

	"src.sourcegraph.com/sourcegraph/app/assets"
	"src.sourcegraph.com/sourcegraph/app/router"
	"src.sourcegraph.com/sourcegraph/conf"
	"src.sourcegraph.com/sourcegraph/util/httputil/httpctx"
)

func robotsTxt(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.FromRequest(r)

	w.Header().Set("Content-Type", "text/plain")
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "User-agent: *")
	if os.Getenv("ROBOTS_TXT_ALLOW") != "" {
		// Disallow user profiles only. On sourcegraph.com user
		// profiles list github repos, which means crawlers will find
		// repos not linked to anywhere else on the web and trigger a
		// clone.
		fmt.Fprintln(&buf, "Disallow: /~")
	} else {
		fmt.Fprintln(&buf, "Disallow: /")
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "Sitemap:", conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.SitemapIndex)).String())
	buf.WriteTo(w)
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, assets.URL("/img/favicon.png").String(), http.StatusMovedPermanently)
}
