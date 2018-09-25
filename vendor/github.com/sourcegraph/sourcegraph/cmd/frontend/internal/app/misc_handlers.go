package app

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/app/assetsutil"
	"github.com/sourcegraph/sourcegraph/pkg/env"
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

	} else {
		fmt.Fprintln(&buf, "Disallow: /")
	}
	fmt.Fprintln(&buf)
	buf.WriteTo(w)
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, assetsutil.URL("/img/favicon.png").String(), http.StatusMovedPermanently)
}
