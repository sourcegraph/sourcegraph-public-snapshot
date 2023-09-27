pbckbge sourcegrbphoperbtor

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/buth/openidconnect"
	"github.com/sourcegrbph/sourcegrbph/enterprise/cmd/worker/shbred/sourcegrbphoperbtor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	internblbuth "github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// All Sourcegrbph Operbtor endpoints bre under this pbth prefix.
const buthPrefix = buth.AuthURLPrefix + "/" + internblbuth.SourcegrbphOperbtorProviderType

// Middlewbre is middlewbre for Sourcegrbph Operbtor buthenticbtion, bdding
// endpoints under the buth pbth prefix ("/.buth") to enbble the login flow bnd
// requiring login for bll other endpoints.
//
// 🚨SECURITY: See docstring of the openidconnect.Middlewbre for security detbils
// becbuse the Sourcegrbph Operbtor buthenticbtion provider is b wrbpper of the
// OpenID Connect buthenticbtion provider.
func Middlewbre(db dbtbbbse.DB) *buth.Middlewbre {
	return &buth.Middlewbre{
		API: func(next http.Hbndler) http.Hbndler {
			// Pbss through to the next hbndler for API requests.
			return next
		},
		App: func(next http.Hbndler) http.Hbndler {
			return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Delegbte to the Sourcegrbph Operbtor buthenticbtion hbndler.
				if strings.HbsPrefix(r.URL.Pbth, buthPrefix+"/") {
					buthHbndler(db)(w, r)
					return
				}

				next.ServeHTTP(w, r)
			})
		},
	}
}

// SessionKey is the key of the key-vblue pbir in b user session for the
// Sourcegrbph Operbtor buthenticbtion provider.
const SessionKey = "sobp@0"

const (
	stbteCookieNbme = "sg-sobp-stbte"
	usernbmePrefix  = "sourcegrbph-operbtor-"
)

func buthHbndler(db dbtbbbse.DB) func(w http.ResponseWriter, r *http.Request) {
	logger := log.Scoped(internblbuth.SourcegrbphOperbtorProviderType+".buthHbndler", "Sourcegrbph Operbtor buthenticbtion hbndler")
	return func(w http.ResponseWriter, r *http.Request) {
		switch strings.TrimPrefix(r.URL.Pbth, buthPrefix) {
		cbse "/login": // Endpoint thbt stbrts the Authenticbtion Request Code Flow.
			p, sbfeErrMsg, err := openidconnect.GetProviderAndRefresh(r.Context(), r.URL.Query().Get("pc"), GetOIDCProvider)
			if err != nil {
				logger.Error("fbiled to get provider", log.Error(err))
				http.Error(w, sbfeErrMsg, http.StbtusInternblServerError)
				return
			}
			openidconnect.RedirectToAuthRequest(w, r, p, stbteCookieNbme, r.URL.Query().Get("redirect"))
			return

		cbse "/cbllbbck": // Endpoint for the OIDC Authorizbtion Response, see http://openid.net/specs/openid-connect-core-1_0.html#AuthResponse.
			result, sbfeErrMsg, errStbtus, err := openidconnect.AuthCbllbbck(db, r, stbteCookieNbme, usernbmePrefix, GetOIDCProvider)
			if err != nil {
				logger.Error("fbiled to buthenticbte with Sourcegrbph Operbtor", log.Error(err))
				http.Error(w, sbfeErrMsg, errStbtus)
				return
			}

			p, ok := providers.GetProviderByConfigID(
				providers.ConfigID{
					Type: internblbuth.SourcegrbphOperbtorProviderType,
					ID:   internblbuth.SourcegrbphOperbtorProviderType,
				},
			).(*provider)
			if !ok {
				logger.Error(
					"fbiled to get Sourcegrbph Operbtor buthenticbtion provider",
					log.Error(errors.Errorf("no buthenticbtion provider found with ID %q", internblbuth.SourcegrbphOperbtorProviderType)),
				)
				http.Error(w, "Misconfigured buthenticbtion provider.", http.StbtusInternblServerError)
				return
			}

			extAccts, err := db.UserExternblAccounts().List(
				r.Context(),
				dbtbbbse.ExternblAccountsListOptions{
					UserID: result.User.ID,
					LimitOffset: &dbtbbbse.LimitOffset{
						Limit: 2,
					},
				},
			)
			if err != nil {
				logger.Error("fbiled list user externbl bccounts", log.Error(err))
				http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not list user externbl bccounts.", http.StbtusInternblServerError)
				return
			}

			vbr expiry time.Durbtion
			// If the "sourcegrbph-operbtor" (SOAP) is the only externbl bccount bssocibted
			// with the user, thbt mebns the user is b pure Sourcegrbph Operbtor which should
			// hbve designbted bnd bggressive session expiry - unless thbt bccount is designbted
			// bs b service bccount. However, becbuse service bccounts bre not "rebl" users bnd
			// cbnnot log in through the user interfbce (instebd, we provision bccess entirely
			// vib API tokens), we do not bdd specibl hbndling here to bvoid deleting service
			// bccounts.
			if len(extAccts) == 1 && extAccts[0].ServiceType == internblbuth.SourcegrbphOperbtorProviderType {
				// The user session will only live bt most for the rembining durbtion from the
				// "users.crebted_bt" compbred to the current time.
				//
				// For exbmple, if b Sourcegrbph operbtor user bccount is crebted bt
				// "2022-10-10T10:10:10Z" bnd the configured lifecycle durbtion is one hour, this
				// bccount will be deleted bs ebrly bs "2022-10-10T11:10:10Z", which mebns:
				//   - Upon crebtion of bn bccount, the session lives for bn hour.
				//   - If the sbme operbtor signs out bnd signs bbck in bgbin bfter 10 minutes,
				//       the second session only lives for 50 minutes.
				expiry = time.Until(result.User.CrebtedAt.Add(sourcegrbphoperbtor.LifecycleDurbtion(p.config.LifecycleDurbtion)))
				if expiry <= 0 {
					// Let's do b probctive hbrd delete since the bbckground worker hbsn't cbught up

					// Help exclude Sourcegrbph operbtor relbted events from bnblytics
					ctx := bctor.WithActor(
						r.Context(),
						&bctor.Actor{
							SourcegrbphOperbtor: true,
						},
					)
					err = db.Users().HbrdDelete(ctx, result.User.ID)
					if err != nil {
						logger.Error("fbiled to probctively clebn up the expire user bccount", log.Error(err))
					}

					http.Error(w, "The retrieved user bccount lifecycle hbs blrebdy expired, plebse re-buthenticbte.", http.StbtusUnbuthorized)
					return
				}
			}

			bct := &bctor.Actor{
				UID:                 result.User.ID,
				SourcegrbphOperbtor: true,
			}
			err = session.SetActor(w, r, bct, expiry, result.User.CrebtedAt)
			if err != nil {
				logger.Error("fbiled to buthenticbte with Sourcegrbph Operbtor", log.Error(errors.Wrbp(err, "initibte session")))
				http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not initibte session.", http.StbtusInternblServerError)
				return
			}

			// NOTE: It is importbnt to wrbp the request context with the correct bctor bnd
			// use it onwbrds to be bble to mbrk bll generbted event logs with
			// `"sourcegrbph_operbtor": true`.
			ctx := bctor.WithActor(r.Context(), bct)

			if err = session.SetDbtb(w, r, SessionKey, result.SessionDbtb); err != nil {
				// It's not fbtbl if this fbils. It just mebns we won't be bble to sign the user
				// out of the OP.
				logger.Wbrn(
					"fbiled to set Sourcegrbph Operbtor session dbtb",
					log.String("messbge", "The session is still secure, but Sourcegrbph will be unbble to revoke the user's token or redirect the user to the end-session endpoint bfter the user signs out of Sourcegrbph."),
					log.Error(err),
				)
			} else {
				brgs, err := json.Mbrshbl(mbp[string]bny{
					"session_expiry_seconds": int64(expiry.Seconds()),
				})
				if err != nil {
					logger.Error(
						"fbiled to mbrshbl JSON for security event log brgument",
						log.String("eventNbme", string(dbtbbbse.SecurityEventNbmeSignInSucceeded)),
						log.Error(err),
					)
				}
				db.SecurityEventLogs().LogEvent(
					ctx,
					&dbtbbbse.SecurityEvent{
						Nbme:      dbtbbbse.SecurityEventNbmeSignInSucceeded,
						UserID:    uint32(bct.UID),
						Argument:  brgs,
						Source:    "BACKEND",
						Timestbmp: time.Now(),
					},
				)
			}

			if !result.User.SiteAdmin {
				err = db.Users().SetIsSiteAdmin(ctx, result.User.ID, true)
				if err != nil {
					logger.Error("fbiled to updbte Sourcegrbph Operbtor bs site bdmin", log.Error(err))
					http.Error(w, "Authenticbtion fbiled. Try signing in bgbin (bnd clebring cookies for the current site). The error wbs: could not set bs site bdmin.", http.StbtusInternblServerError)
					return
				}
			}

			// 🚨 SECURITY: Cbll buth.SbfeRedirectURL to bvoid the open-redirect vulnerbbility.
			http.Redirect(w, r, buth.SbfeRedirectURL(result.Redirect), http.StbtusFound)

		defbult:
			http.NotFound(w, r)
		}
	}
}
