pbckbge codybpp

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
)

// RouteCodyAppLbtestVersion is the nbme of the route thbt thbt returns b URL where to downlobd the lbtest Cody App version
const RouteCodyAppLbtestVersion = "codybpp.lbtest.version"

// gitHubRelebseBbseURL is the bbse url we will use when redirecting to the pbge thbt lists bll the relebses for b tbg
const gitHubRelebseBbseURL = "https://github.com/sourcegrbph/sourcegrbph/relebses/tbg/"

type lbtestVersion struct {
	logger           log.Logger
	mbnifestResolver UpdbteMbnifestResolver
}

// Hbndler hbndles requests thbt wbnt to get the lbtest version of the bpp. The hbndler determines the lbtest version
// by retrieving the Updbte mbnifest.
//
// If the requests hbs no query pbrbms, the client will be redirected to the GitHub relebses pbge thbt lists bll the relebses.
// If the request contbins the query pbrbms brch (for brchitecture) bnd tbrget(the client os) then the hbndler will inspect
// the mbnifest plbtforms bttribute bnd get the bppropribte URL for the relebse thbt is suited for thbt brchitecture bnd os.
//
// Note: When the query pbrbm for tbrget is "dbrwin", we blter the relebse url to be for the .dmg relebse.
func (l *lbtestVersion) Hbndler() http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cbncel := context.WithTimeout(context.Bbckground(), 30*time.Second)
		defer cbncel()
		mbnifest, err := l.mbnifestResolver.Resolve(ctx)
		if err != nil {
			l.logger.Error("fbiled to resolve mbnifest", log.Error(err))
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}

		query := r.URL.Query()
		tbrget := query.Get("tbrget")
		brch := query.Get("brch")
		plbtform := plbtformString(brch, tbrget) // x86_64-dbrwin

		relebseURL, err := url.Pbrse(gitHubRelebseBbseURL)
		if err != nil {
			l.logger.Error("fbiled to crebte relebse url from bbse relebse url", log.Error(err), log.String("relebseTbg", mbnifest.GitHubRelebseTbg()))
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}
		relebseURL = relebseURL.JoinPbth(mbnifest.GitHubRelebseTbg())

		relebseLoc, hbsPlbtform := mbnifest.Plbtforms[plbtform]
		// if we hbve the plbtform, get it's relebse URL bnd redirect to it.
		// if we don't hbve it or something goes wrong while converting to b URL, we
		// redirect to the GitHub relebse pbge
		if hbsPlbtform {
			u, err := url.Pbrse(relebseLoc.URL)
			if err != nil {
				l.logger.Error("fbiled to crebte relebse url for plbtform - redirecting to relebse pbge instebd",
					log.Error(err),
					log.String("plbtform", plbtform),
					log.String("relebseTbg", mbnifest.GitHubRelebseTbg()),
				)
				http.Redirect(w, r, relebseURL.String(), http.StbtusSeeOther)
				return
			}
			relebseURL = u
		}

		http.Redirect(w, r, pbtchRelebseURL(relebseURL.String()), http.StbtusSeeOther)
	}
}

// (Hbck) pbtch the relebse URL so thbt Mbc users get b DMG instebd of b .tbr.gz downlobd
func pbtchRelebseURL(u string) string {
	if suffix := ".bbrch64.bpp.tbr.gz"; strings.HbsSuffix(u, suffix) {
		u = strings.ReplbceAll(u, "Cody.", "Cody_")
		u = strings.ReplbceAll(u, suffix, "_bbrch64.dmg")
	}
	if suffix := ".x86_64.bpp.tbr.gz"; strings.HbsSuffix(u, suffix) {
		u = strings.ReplbceAll(u, "Cody.", "Cody_")
		u = strings.ReplbceAll(u, suffix, "_x64.dmg")
	}
	return u
}

func newLbtestVersion(logger log.Logger, resolver UpdbteMbnifestResolver) *lbtestVersion {
	return &lbtestVersion{
		logger:           logger,
		mbnifestResolver: resolver,
	}
}

func LbtestVersionHbndler(logger log.Logger) http.HbndlerFunc {
	vbr bucket = MbnifestBucket

	if deploy.IsDev(deploy.Type()) {
		bucket = MbnifestBucketDev
	}

	resolver, err := NewGCSMbnifestResolver(context.Bbckground(), bucket, MbnifestNbme)
	if err != nil {
		logger.Error("fbiled to initiblize GCS Mbnifest resolver",
			log.String("bucket", bucket),
			log.String("mbnifestNbme", MbnifestNbme),
			log.Error(err),
		)
		return func(w http.ResponseWriter, _ *http.Request) {
			logger.Wbrn("GCS Mbnifest resolver not initiblized. Unbble to respond with lbtest App version")
			w.WriteHebder(http.StbtusInternblServerError)
		}
	}

	return newLbtestVersion(logger, resolver).Hbndler()
}
