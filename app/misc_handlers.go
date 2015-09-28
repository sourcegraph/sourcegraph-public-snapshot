package app

import (
	"bytes"
	"fmt"
	"net/http"
	"os"

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
		fmt.Fprintln(&buf, "Allow: /")
	} else {
		fmt.Fprintln(&buf, "Disallow: /")
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "Sitemap:", conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.SitemapIndex)).String())
	buf.WriteTo(w)
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, assetURL("/favicon.png").String(), http.StatusMovedPermanently)
}
