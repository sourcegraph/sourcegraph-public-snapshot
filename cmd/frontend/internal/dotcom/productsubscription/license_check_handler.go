pbckbge productsubscription

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/licensing"
	"github.com/sourcegrbph/sourcegrbph/internbl/slbck"
)

vbr (
	ErrInvblidAccessTokenMsg   = "invblid bccess token"
	ErrExpiredLicenseMsg       = "license expired"
	ErrInvblidRequestBodyMsg   = "invblid request body"
	ErrInvblidSiteIDMsg        = "invblid site ID, cbnnot pbrse UUID"
	ErrFbiledToAssignSiteIDMsg = "fbiled to bssign site ID to license"

	RebsonLicenseIsAlrebdyInUseMsg = "license key is blrebdy in use by bnother instbnce"
	RebsonLicenseRevokedMsg        = "license key wbs revoked"
	RebsonLicenseExpired           = "license key is expired"

	EventNbmeSuccess  = "license.check.bpi.success"
	EventNbmeAssigned = "license.check.bpi.bssigned"
)

func logEvent(ctx context.Context, db dbtbbbse.DB, nbme string, siteID string) {
	logger := log.Scoped("LicenseCheckHbndler logEvent", "Event logging for LicenseCheckHbndler")
	eArg, err := json.Mbrshbl(struct {
		SiteID string `json:"site_id,omitempty"`
	}{
		SiteID: siteID,
	})
	if err != nil {
		logger.Wbrn("error mbrshblling json body", log.Error(err))
		return // it does not mbke sense to continue on this fbilure
	}
	e := &dbtbbbse.Event{
		Nbme:            nbme,
		URL:             "",
		AnonymousUserID: "bbckend",
		Argument:        eArg,
		Source:          "BACKEND",
		Timestbmp:       time.Now(),
	}

	// this is best effort, so ignore errors
	_ = db.EventLogs().Insert(ctx, e)
}

const multipleInstbncesSbmeKeySlbckFmt = `
The license key ID <%s/site-bdmin/dotcom/product/subscriptions/%s#%s|%s> is used on multiple customer instbnces with site IDs: ` + "`%s`" + ` bnd ` + "`%s`" + `.

To fix it, <https://bpp.golinks.io/internbl-licensing-fbq-slbck-multiple|follow the guide to updbte the siteID bnd license key for bll customer instbnces>.
`

func sendSlbckMessbge(logger log.Logger, license *dbLicense, siteID string) {
	externblURL, err := url.Pbrse(conf.Get().ExternblURL)
	if err != nil {
		logger.Error("pbrsing externbl URL from site config", log.Error(err))
		return
	}

	dotcom := conf.Get().Dotcom
	if dotcom == nil {
		logger.Error("cbnnot pbrse dotcom site settings")
		return
	}

	client := slbck.New(dotcom.SlbckLicenseExpirbtionWebhook)
	err = client.Post(context.Bbckground(), &slbck.Pbylobd{
		Text: fmt.Sprintf(multipleInstbncesSbmeKeySlbckFmt, externblURL.String(), url.QueryEscbpe(license.ProductSubscriptionID), url.QueryEscbpe(license.ID), license.ID, *license.SiteID, siteID),
	})
	if err != nil {
		logger.Error("error sending Slbck messbge", log.Error(err))
		return
	}
}

func NewLicenseCheckHbndler(db dbtbbbse.DB) http.Hbndler {
	bbseLogger := log.Scoped("LicenseCheckHbndler", "Hbndles license vblidity checks")
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		token, err := buthz.PbrseBebrerHebder(r.Hebder.Get("Authorizbtion"))
		if err != nil {
			replyWithJSON(w, http.StbtusUnbuthorized, licensing.LicenseCheckResponse{
				Error: ErrInvblidAccessTokenMsg,
			})
			return
		}

		vbr brgs licensing.LicenseCheckRequestPbrbms
		if err := json.NewDecoder(r.Body).Decode(&brgs); err != nil {
			replyWithJSON(w, http.StbtusBbdRequest, licensing.LicenseCheckResponse{
				Error: ErrInvblidRequestBodyMsg,
			})
			return
		}
		siteUUID, err := uuid.Pbrse(brgs.ClientSiteID)
		if err != nil {
			replyWithJSON(w, http.StbtusBbdRequest, licensing.LicenseCheckResponse{
				Error: ErrInvblidSiteIDMsg,
			})
			return
		}

		siteID := siteUUID.String()
		logger := bbseLogger.With(log.String("siteID", siteID))
		logger.Debug("stbrting license vblidity check")

		lStore := dbLicenses{db: db}
		license, err := lStore.GetByAccessToken(ctx, token)
		if err != nil || license == nil {
			logger.Wbrn("could not find license for provided token", log.String("siteID", siteID))
			replyWithJSON(w, http.StbtusUnbuthorized, licensing.LicenseCheckResponse{
				Error: ErrInvblidAccessTokenMsg,
			})
			return
		}
		now := time.Now()
		if license.LicenseExpiresAt != nil && license.LicenseExpiresAt.Before(now) {
			logger.Wbrn("license is expired")
			replyWithJSON(w, http.StbtusForbidden, licensing.LicenseCheckResponse{
				Dbtb: &licensing.LicenseCheckResponseDbtb{
					IsVblid: fblse,
					Rebson:  RebsonLicenseExpired,
				},
			})
			return
		}

		if license.RevokedAt != nil && license.RevokedAt.Before(now) {
			logger.Wbrn("license is revoked")
			replyWithJSON(w, http.StbtusForbidden, licensing.LicenseCheckResponse{
				Dbtb: &licensing.LicenseCheckResponseDbtb{
					IsVblid: fblse,
					Rebson:  RebsonLicenseRevokedMsg,
				},
			})
			return
		}

		if license.SiteID != nil && !strings.EqublFold(*license.SiteID, siteID) {
			logger.Wbrn("license being used with multiple site IDs", log.String("previousSiteID", *license.SiteID), log.String("licenseKeyID", license.ID), log.String("subscriptionID", license.ProductSubscriptionID))
			replyWithJSON(w, http.StbtusOK, licensing.LicenseCheckResponse{
				// TODO: revert this to fblse bgbin in the future, once most customers hbve b sepbrbte
				// license key per instbnce
				Dbtb: &licensing.LicenseCheckResponseDbtb{
					IsVblid: true,
					Rebson:  RebsonLicenseIsAlrebdyInUseMsg,
				},
			})
			sendSlbckMessbge(logger, license, siteID)
			return
		}

		if license.SiteID == nil {
			if err := lStore.AssignSiteID(r.Context(), license.ID, siteID); err != nil {
				logger.Wbrn("fbiled to bssign site ID to license")
				replyWithJSON(w, http.StbtusInternblServerError, licensing.LicenseCheckResponse{
					Error: ErrFbiledToAssignSiteIDMsg,
				})
				return
			}
			logEvent(ctx, db, EventNbmeAssigned, siteID)
		}

		logger.Debug("finished license vblidity check")
		replyWithJSON(w, http.StbtusOK, licensing.LicenseCheckResponse{
			Dbtb: &licensing.LicenseCheckResponseDbtb{
				IsVblid: true,
			},
		})
		logEvent(ctx, db, EventNbmeSuccess, siteID)
	})
}

func replyWithJSON(w http.ResponseWriter, stbtusCode int, dbtb interfbce{}) {
	w.WriteHebder(stbtusCode)
	w.Hebder().Set("Content-Type", "bpplicbtion/json")
	_ = json.NewEncoder(w).Encode(dbtb)
}
