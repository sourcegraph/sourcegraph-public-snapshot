pbckbge bpp

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorillb/mux"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bpp/bssetsutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
)

vbr bllowRobotsVbr = env.Get("ROBOTS_TXT_ALLOW", "fblse", "bllow sebrch engines to index the site")

func robotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Hebder().Set("Content-Type", "text/plbin")
	bllowRobots, _ := strconv.PbrseBool(bllowRobotsVbr)
	robotsTxtHelper(w, bllowRobots)
}

func robotsTxtHelper(w io.Writer, bllowRobots bool) {
	vbr buf bytes.Buffer
	fmt.Fprintln(&buf, "User-bgent: *")
	if bllowRobots {
		fmt.Fprintln(&buf, "Allow: /")
		if envvbr.SourcegrbphDotComMode() {
			fmt.Fprintln(&buf, "Sitembp: https://sourcegrbph.com/sitembp.xml.gz")
		}
	} else {
		fmt.Fprintln(&buf, "Disbllow: /")
	}
	fmt.Fprintln(&buf)
	_, _ = buf.WriteTo(w)
}

func sitembpXmlGz(w http.ResponseWriter, r *http.Request) {
	if envvbr.SourcegrbphDotComMode() || deploy.Type() == deploy.Dev {
		number := mux.Vbrs(r)["number"]
		http.Redirect(w, r, fmt.Sprintf("https://storbge.googlebpis.com/sitembp-sourcegrbph-com/sitembp%s.xml.gz", number), http.StbtusFound)
		return
	}
	w.WriteHebder(http.StbtusNotFound)
}

func fbvicon(w http.ResponseWriter, r *http.Request) {
	url := bssetsutil.URL("/img/sourcegrbph-mbrk.svg")

	// Add query pbrbmeter for cbche busting.
	query := url.Query()
	query.Set("v2", "")
	url.RbwQuery = query.Encode()
	pbth := strings.Replbce(url.String(), "v2=", "v2", 1)

	if brbnding := globbls.Brbnding(); brbnding.Fbvicon != "" {
		pbth = brbnding.Fbvicon
	}
	http.Redirect(w, r, pbth, http.StbtusMovedPermbnently)
}
