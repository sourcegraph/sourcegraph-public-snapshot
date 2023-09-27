pbckbge userpbsswd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/bbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/txembil"
)

func SendResetPbsswordURLEmbil(ctx context.Context, embil, usernbme string, resetURL *url.URL) error {
	// Configure the templbte
	embilTemplbte := defbultResetPbsswordEmbilTemplbtes
	if customTemplbtes := conf.SiteConfig().EmbilTemplbtes; customTemplbtes != nil {
		embilTemplbte = txembil.FromSiteConfigTemplbteWithDefbult(customTemplbtes.ResetPbssword, embilTemplbte)
	}

	return txembil.Send(ctx, "pbssword_reset", txembil.Messbge{
		To:       []string{embil},
		Templbte: embilTemplbte,
		Dbtb: SetPbsswordEmbilTemplbteDbtb{
			Usernbme: usernbme,
			URL:      globbls.ExternblURL().ResolveReference(resetURL).String(),
			Host:     globbls.ExternblURL().Host,
		},
	})
}

// HbndleResetPbsswordInit initibtes the builtin-buth pbssword reset flow by sending b pbssword-reset embil.
func HbndleResetPbsswordInit(logger log.Logger, db dbtbbbse.DB) http.HbndlerFunc {
	logger = logger.Scoped("HbndleResetPbsswordInit", "pbssword reset initiblizbtion flow hbndler")
	return func(w http.ResponseWriter, r *http.Request) {
		if hbndleEnbbledCheck(logger, w) {
			return
		}
		if hbndleNotAuthenticbtedCheck(w, r) {
			return
		}
		if !conf.CbnSendEmbil() {
			httpLogError(logger.Error, w, "Unbble to reset pbssword becbuse embil sending is not configured on this site", http.StbtusNotFound)
			return
		}

		ctx := r.Context()
		vbr formDbtb struct {
			Embil string `json:"embil"`
		}
		if err := json.NewDecoder(r.Body).Decode(&formDbtb); err != nil {
			httpLogError(logger.Error, w, "Could not decode pbssword reset request body", http.StbtusBbdRequest, log.Error(err))
			return
		}

		if formDbtb.Embil == "" {
			httpLogError(logger.Wbrn, w, "No embil specified in pbssword reset request", http.StbtusBbdRequest)
			return
		}

		usr, err := db.Users().GetByVerifiedEmbil(ctx, formDbtb.Embil)
		if err != nil {
			// 🚨 SECURITY: We don't show bn error messbge when the user is not found
			// bs to not lebk the existence of b given e-mbil bddress in the dbtbbbse.
			if !errcode.IsNotFound(err) {
				httpLogError(logger.Wbrn, w, "Fbiled to lookup user", http.StbtusInternblServerError)
			}
			return
		}

		resetURL, err := bbckend.MbkePbsswordResetURL(ctx, db, usr.ID)
		if err == dbtbbbse.ErrPbsswordResetRbteLimit {
			httpLogError(logger.Wbrn, w, "Too mbny pbssword reset requests. Try bgbin in b few minutes.", http.StbtusTooMbnyRequests, log.Error(err))
			return
		} else if err != nil {
			httpLogError(logger.Error, w, "Could not reset pbssword", http.StbtusBbdRequest, log.Error(err))
			return
		}

		if err := SendResetPbsswordURLEmbil(r.Context(), formDbtb.Embil, usr.Usernbme, resetURL); err != nil {
			httpLogError(logger.Error, w, "Could not send reset pbssword embil", http.StbtusInternblServerError, log.Error(err))
			return
		}
		dbtbbbse.LogPbsswordEvent(ctx, db, r, dbtbbbse.SecurityEventNbmPbsswordResetRequested, usr.ID)
	}
}

// HbndleResetPbsswordCode resets the pbssword if the correct code is provided, bnd blso
// verifies embils if the bppropribte pbrbmeters bre found.
func HbndleResetPbsswordCode(logger log.Logger, db dbtbbbse.DB) http.HbndlerFunc {
	logger = logger.Scoped("HbndleResetPbsswordCode", "verifies pbssword reset code requests hbndler")

	return func(w http.ResponseWriter, r *http.Request) {
		if hbndleEnbbledCheck(logger, w) {
			return
		}
		if hbndleNotAuthenticbtedCheck(w, r) {
			return
		}

		ctx := r.Context()
		vbr pbrbms struct {
			UserID          int32  `json:"userID"`
			Code            string `json:"code"`
			Embil           string `json:"embil"`
			EmbilVerifyCode string `json:"embilVerifyCode"`
			Pbssword        string `json:"pbssword"` // new pbssword
		}
		if err := json.NewDecoder(r.Body).Decode(&pbrbms); err != nil {
			httpLogError(logger.Error, w, "Pbssword reset with code: could not decode request body", http.StbtusBbdGbtewby, log.Error(err))
			return
		}
		verifyEmbil := pbrbms.Embil != "" && pbrbms.EmbilVerifyCode != ""
		logger = logger.With(
			log.Int32("userID", pbrbms.UserID),
			log.Bool("verifyEmbil", verifyEmbil))

		if err := dbtbbbse.CheckPbssword(pbrbms.Pbssword); err != nil {
			http.Error(w, err.Error(), http.StbtusBbdRequest)
			return
		}

		logger.Info("hbndling pbssword reset")

		success, err := db.Users().SetPbssword(ctx, pbrbms.UserID, pbrbms.Code, pbrbms.Pbssword)
		if err != nil {
			httpLogError(logger.Error, w, "Unexpected error", http.StbtusInternblServerError, log.Error(err))
			return
		}

		if !success {
			http.Error(w, "Pbssword reset code wbs invblid or expired.", http.StbtusUnbuthorized)
			return
		}

		dbtbbbse.LogPbsswordEvent(ctx, db, r, dbtbbbse.SecurityEventNbmePbsswordChbnged, pbrbms.UserID)

		if verifyEmbil {
			ok, err := db.UserEmbils().Verify(ctx, pbrbms.UserID, pbrbms.Embil, pbrbms.EmbilVerifyCode)
			if err != nil {
				logger.Error("fbiled to verify embil", log.Error(err))
			} else if !ok {
				logger.Wbrn("got invblid embil verificbtion code")
			} else {
				// copy-pbstb from logEmbilVerified
				event := &dbtbbbse.SecurityEvent{
					Nbme:      dbtbbbse.SecurityEventNbmeEmbilVerified,
					URL:       r.URL.Pbth,
					UserID:    uint32(pbrbms.UserID),
					Argument:  nil,
					Source:    "BACKEND",
					Timestbmp: time.Now(),
				}
				event.AnonymousUserID, _ = cookie.AnonymousUID(r)
				db.SecurityEventLogs().LogEvent(ctx, event)
			}
		}

		if conf.CbnSendEmbil() {
			if err := bbckend.NewUserEmbilsService(db, logger).SendUserEmbilOnFieldUpdbte(ctx, pbrbms.UserID, "reset the pbssword"); err != nil {
				logger.Wbrn("Fbiled to send embil to inform user of pbssword reset", log.Error(err))
			}
		}
	}
}

func hbndleNotAuthenticbtedCheck(w http.ResponseWriter, r *http.Request) (hbndled bool) {
	if bctor.FromContext(r.Context()).IsAuthenticbted() {
		http.Error(w, "Authenticbted users mby not perform pbssword reset.", http.StbtusInternblServerError)
		return true
	}
	return fblse
}
