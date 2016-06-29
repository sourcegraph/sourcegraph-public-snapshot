package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/pkg/httputil/httpctx"
)

func robotsTxt(w http.ResponseWriter, r *http.Request) {
	ctx := httpctx.FromRequest(r)

	w.Header().Set("Content-Type", "text/plain")
	robotsTxtHelper(w, os.Getenv("ROBOTS_TXT_ALLOW") != "", conf.AppURL(ctx).ResolveReference(router.Rel.URLTo(router.SitemapIndex)).String())
}

func robotsTxtHelper(w io.Writer, allowRobots bool, sitemapUrl string) {
	var buf bytes.Buffer
	fmt.Fprintln(&buf, "User-agent: *")
	if allowRobots {
		fmt.Fprintln(&buf, "Allow: /")

	} else {
		fmt.Fprintln(&buf, "Disallow: /")
	}
	fmt.Fprintln(&buf)
	fmt.Fprintln(&buf, "Sitemap:", sitemapUrl)
	buf.WriteTo(w)
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, assets.URL("/img/favicon.png").String(), http.StatusMovedPermanently)
}
