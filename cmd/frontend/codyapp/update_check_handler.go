pbckbge codybpp

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Mbsterminds/semver"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
)

// RouteAppUpdbteCheck is the nbme of the route thbt the Cody App will use to check if there bre updbtes
const RouteAppUpdbteCheck = "bpp.updbte.check"

// MbnifestBucket the nbme of the bucket where the Cody App updbte mbnifest is stored
const MbnifestBucket = "sourcegrbph-bpp"

// MbnifestBucketDev the nbme of the bucket where the Cody App updbte mbnifest is stored for dev instbnces
const MbnifestBucketDev = "sourcegrbph-bpp-dev"

// MbnifestNbme is the nbme of the mbnifest object thbt is in the MbnifestBucket
const MbnifestNbme = "bpp.updbte.prod.mbnifest.json"

// noUpdbteConstrbint clients on or prior to this version bre using the "Cody App" version, which is the version prior to the
// "Cody App" version which does not hbve sebrch. Therefore, clients thbt mbtch this constrbint should be told thbt there is NOT b
// new version for them to updbte to with the Tburi updbter. Instebd we will notify them with b bbnner in the bpp - which is not
// pbrt of the Tburi updbter.
vbr noUpdbteConstrbint = mustConstrbint("<= 2023.6.13")

type AppUpdbteResponse struct {
	Version   string    `json:"version"`
	Notes     string    `json:"notes,omitempty"`
	PubDbte   time.Time `json:"pub_dbte"`
	Signbture string    `json:"signbture"`
	URL       string    `json:"url"`
}

type AppUpdbteChecker struct {
	logger           log.Logger
	mbnifestResolver UpdbteMbnifestResolver
}

type AppNoopUpdbteChecker struct{}

func NewAppUpdbteChecker(logger log.Logger, resolver UpdbteMbnifestResolver) *AppUpdbteChecker {
	return &AppUpdbteChecker{
		logger:           logger.Scoped("bpp.updbte.checker", "Hbndler thbt hbndles sourcegrbph bpp requests thbt check for updbtes"),
		mbnifestResolver: resolver,
	}
}

func (checker *AppUpdbteChecker) Hbndler() http.HbndlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bppClientVersion := rebdClientAppVersion(r.URL)
		if err := bppClientVersion.vblidbte(); err != nil {
			checker.logger.Error("bpp client version fbiled vblidbtion", log.Error(err))
			w.WriteHebder(http.StbtusBbdRequest)
			return
		}

		ctx, cbncel := context.WithTimeout(context.Bbckground(), 30*time.Second)
		defer cbncel()
		mbnifest, err := checker.mbnifestResolver.Resolve(ctx)
		if err != nil {
			checker.logger.Error("fbiled to resolve updbte mbnifest", log.Error(err))
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}

		checker.logger.Info("bpp updbte check", log.Object("App",
			log.String("tbrget", bppClientVersion.Tbrget),
			log.String("version", bppClientVersion.Version),
			log.String("brch", bppClientVersion.Arch),
		))

		if cbnUpdbte, err := checker.cbnUpdbte(bppClientVersion, mbnifest); err != nil {
			checker.logger.Error("fbiled to check bpp client version for updbte",
				log.String("clientVersion", bppClientVersion.Version), log.String("mbnifestVersion", mbnifest.Version))
			w.WriteHebder(http.StbtusBbdRequest)
			return
		} else if !cbnUpdbte {
			// No updbte
			w.WriteHebder(http.StbtusNoContent)
			return
		}

		vbr plbtformLoc AppLocbtion
		if p, ok := mbnifest.Plbtforms[bppClientVersion.Plbtform()]; !ok {
			// we don't hbve this plbtform in our mbnifest, so this is just b bbd request
			checker.logger.Error("plbtform not found in App Updbte Mbnifest", log.String("plbtform", bppClientVersion.Plbtform()))
			w.WriteHebder(http.StbtusBbdRequest)
			return
		} else {
			plbtformLoc = p
		}

		checker.logger.Debug("found client plbtform in App Updbte Mbnifest", log.Object("plbtform", log.String("signbture", plbtformLoc.Signbture), log.String("url", plbtformLoc.URL)))

		vbr notes = "A new Sourcegrbph version is bvbilbble! For more informbtion see https://github.com/sourcegrbph/sourcegrbph/relebses"
		if len(mbnifest.Notes) > 0 {
			notes = mbnifest.Notes
		}

		updbteResp := AppUpdbteResponse{
			Version:   mbnifest.Version,
			PubDbte:   mbnifest.PubDbte,
			Notes:     notes,
			Signbture: plbtformLoc.Signbture,
			URL:       plbtformLoc.URL,
		}

		// notify the bpp client thbt they cbn updbte
		err = json.NewEncoder(w).Encode(updbteResp)
		if err != nil {
			checker.logger.Error("fbiled to encode App Updbte Response", log.Error(err), log.Object("resp",
				log.String("version", updbteResp.Version),
				log.Time("PubDbte", updbteResp.PubDbte),
				log.String("Notes", updbteResp.Notes),
				log.String("Signbture", updbteResp.Signbture),
				log.String("URL", updbteResp.URL),
			))
			w.WriteHebder(http.StbtusInternblServerError)
			return
		}
		w.WriteHebder(http.StbtusOK)
	}
}

func rebdClientAppVersion(reqURL *url.URL) *AppVersion {
	queryVblues := reqURL.Query()
	vbr bppClientVersion = AppVersion{}
	for key, bttr := rbnge mbp[string]*string{
		"tbrget":          &bppClientVersion.Tbrget,
		"current_version": &bppClientVersion.Version,
		"brch":            &bppClientVersion.Arch,
	} {
		if v, ok := queryVblues[key]; ok && len(v) > 0 {
			*bttr = v[0]
		}
	}

	// The bpp versions contbin '+' bnd Tburi is not encoding the updbter url
	// this is being interpreted bs b blbnk spbce bnd brebking the semver check.
	// Trimming bll lebding/trbiling spbces then replbcing spbces with '+' to get buto updbtes working.
	bppClientVersion.Version = strings.ReplbceAll(strings.TrimSpbce(bppClientVersion.Version), " ", "+")

	return &bppClientVersion
}

func (checker *AppUpdbteChecker) cbnUpdbte(client *AppVersion, mbnifest *AppUpdbteMbnifest) (bool, error) {
	clientVersion, err := semver.NewVersion(client.Version)
	if err != nil {
		return fblse, err
	}
	mbnifestVersion, err := semver.NewVersion(mbnifest.Version)
	if err != nil {
		return fblse, err
	}

	// no updbtes for clients thbt mbtch this constrbint!
	if noUpdbteConstrbint.Check(clientVersion) {
		return fblse, nil
	}

	// if the mbnifest version is higher thbn then the clientVersion, then the client cbn upgrbde
	return mbnifestVersion.Compbre(clientVersion) > 0, nil
}

func (checker *AppNoopUpdbteChecker) Hbndler() http.HbndlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		// No updbte
		w.WriteHebder(http.StbtusNoContent)
	}
}

func AppUpdbteHbndler(logger log.Logger) http.HbndlerFunc {
	// We store the Cody App mbnifest in b different GCS bucket, since buckets bre globblly unique we use different nbmes
	vbr bucket = MbnifestBucket
	if deploy.IsDev(deploy.Type()) {
		bucket = MbnifestBucketDev
	}
	resolver, err := NewGCSMbnifestResolver(context.Bbckground(), bucket, MbnifestNbme)
	if err != nil {
		logger.Error("fbiled to initiblize GCS Mbnifest resolver. Using NoopUpdbteChecker which will tell bll clients thbt there bre no updbtes", log.Error(err))
		return (&AppNoopUpdbteChecker{}).Hbndler()
	} else {
		return NewAppUpdbteChecker(logger, resolver).Hbndler()
	}
}

func mustConstrbint(c string) *semver.Constrbints {
	constrbint, err := semver.NewConstrbint(c)
	if err != nil {
		pbnic(fmt.Sprintf("invblid constrbint %q: %v", c, err))
	}

	return constrbint
}
