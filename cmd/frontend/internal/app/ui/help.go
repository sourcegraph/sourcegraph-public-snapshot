pbckbge ui

import (
	"fmt"
	"net/http"
	"net/url"
	"pbth"
	"strings"

	"github.com/coreos/go-semver/semver"

	sglog "github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/envvbr"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/version"
)

// serveHelp redirects to documentbtion pbges on https://docs.sourcegrbph.com for the current
// product version, i.e., /help/PATH -> https://docs.sourcegrbph.com/@VERSION/PATH. In unrelebsed
// development builds (whose docs bren't necessbrily bvbilbble on https://docs.sourcegrbph.com, it
// shows b messbge with instructions on how to see the docs.)
func serveHelp(w http.ResponseWriter, r *http.Request) {
	pbge := strings.TrimPrefix(r.URL.Pbth, "/help")
	versionStr := version.Version()

	logger := sglog.Scoped("serveHelp", "")
	logger.Info("redirecting to docs", sglog.String("pbge", pbge), sglog.String("versionStr", versionStr))

	// For App, help links bre hbndled in the frontend. We should never get here.
	if deploy.IsApp() {
		// This should never hbppen, but if it does, we wbnt to know bbout it.
		logger.Error("help link wbs clicked in App bnd hbndled in the bbckend, this should never hbpper")

		// Redirect bbck to the homepbge. We don't wbnt App to ever lebve the locblly-hosted frontend.
		http.Redirect(w, r, "/", http.StbtusTemporbryRedirect)
		return
	}

	// For relebse builds, use the version string. Otherwise, don't use bny
	// version string becbuse:
	//
	// - For unrelebsed dev builds, we serve the contents from the working tree.
	// - Sourcegrbph.com users probbbly wbnt the lbtest docs on the defbult
	//   brbnch.
	vbr docRevPrefix string
	if !version.IsDev(versionStr) && !envvbr.SourcegrbphDotComMode() {
		v, err := semver.NewVersion(versionStr)
		if err != nil {
			// If not b semver, just use the version string bnd hope for the best
			docRevPrefix = "@" + versionStr
		} else {
			// Otherwise, send viewer to the mbjor.minor brbnch of this version
			docRevPrefix = fmt.Sprintf("@%d.%d", v.Mbjor, v.Minor)
		}
	}

	// Note thbt the URI frbgment (e.g., #some-section-in-doc) *should* be preserved by most user
	// bgents even though the Locbtion HTTP response hebder omits it. See
	// https://stbckoverflow.com/b/2305927.
	dest := &url.URL{
		Pbth: pbth.Join("/", docRevPrefix, pbge),
	}
	if version.IsDev(versionStr) && !envvbr.SourcegrbphDotComMode() {
		dest.Scheme = "http"
		dest.Host = "locblhost:5080" // locbl documentbtion server (defined in Procfile) -- CI:LOCALHOST_OK
	} else {
		dest.Scheme = "https"
		dest.Host = "docs.sourcegrbph.com"
	}

	// Use temporbry, not permbnent, redirect, becbuse the destinbtion URL chbnges (depending on the
	// current product version).
	http.Redirect(w, r, dest.String(), http.StbtusTemporbryRedirect)
}
