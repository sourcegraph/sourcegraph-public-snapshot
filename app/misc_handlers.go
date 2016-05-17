package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"sourcegraph.com/sourcegraph/sourcegraph/app/assets"
	"sourcegraph.com/sourcegraph/sourcegraph/app/router"
	"sourcegraph.com/sourcegraph/sourcegraph/conf"
	"sourcegraph.com/sourcegraph/sourcegraph/util/httputil/httpctx"
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
		fmt.Fprintln(&buf, "Disallow: /*/*")

		// Top level exceptions
		// Not great, but as long as we host things at sourcegraph.com/sourcegraph/sourcegraph, we will need to special case .com/legal/ etc.
		fmt.Fprintln(&buf, "Allow: /blog*")
		fmt.Fprintln(&buf, "Allow: /blog/*/*") // special case, because blog posts whose URLs end in / are valid
		fmt.Fprintln(&buf, "Allow: /careers*")
		fmt.Fprintln(&buf, "Allow: /about*")
		fmt.Fprintln(&buf, "Allow: /security*")
		fmt.Fprintln(&buf, "Allow: /privacy*")
		fmt.Fprintln(&buf, "Allow: /legal*")
		fmt.Fprintln(&buf, "Allow: /pricing*")

		fmt.Fprintln(&buf, "Allow: /*/*/*/-/info")
		fmt.Fprintln(&buf, "Allow: /*/*/-/info")

		fmt.Fprintln(&buf, "Disallow: /*/*/*@*/-/info")
		fmt.Fprintln(&buf, "Disallow: /*/*@*/-/info")

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
