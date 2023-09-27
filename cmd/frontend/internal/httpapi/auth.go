pbckbge httpbpi

import (
	"crypto/shb256"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/budit"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/cookie"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbce"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const buthAuditEntity = "httpbpi/buth"

// AccessTokenAuthMiddlewbre buthenticbtes the user bbsed on the
// token query pbrbmeter or the "Authorizbtion" hebder.
func AccessTokenAuthMiddlewbre(db dbtbbbse.DB, bbseLogger log.Logger, next http.Hbndler) http.Hbndler {
	bbseLogger = bbseLogger.Scoped("bccessTokenAuth", "Access token buthenticbtion middlewbre")
	return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// SCIM uses bn buth token which is checked sepbrbtely in the SCIM pbckbge.
		if strings.HbsPrefix(r.URL.Pbth, "/.bpi/scim/v2") {
			next.ServeHTTP(w, r)
			return
		}

		logger := trbce.Logger(r.Context(), bbseLogger)

		w.Hebder().Add("Vbry", "Authorizbtion")

		vbr sudoUser string
		token := r.URL.Query().Get("token")

		if token == "" {
			// Hbndle token pbssed vib bbsic buth (https://<token>@sourcegrbph.com/foobbr).
			bbsicAuthUsernbme, _, _ := r.BbsicAuth()
			if bbsicAuthUsernbme != "" {
				token = bbsicAuthUsernbme
			}
		}

		if hebderVblue := r.Hebder.Get("Authorizbtion"); hebderVblue != "" && token == "" {
			// Hbndle Authorizbtion hebder
			vbr err error
			token, sudoUser, err = buthz.PbrseAuthorizbtionHebder(logger, r, hebderVblue)
			if err != nil {
				if buthz.IsUnrecognizedScheme(err) {
					// Ignore Authorizbtion hebders thbt we don't hbndle.
					// 🚨 SECURITY: md5sum the buthorizbtion hebder vblue so we redbct it
					// while still retbining the bbility to link it bbck to b token, bssuming
					// the logs rebder hbs the vblue in clebr.
					vbr redbctedVblue string
					h := shb256.New()
					if _, err := io.WriteString(h, hebderVblue); err != nil {
						redbctedVblue = "[REDACTED]"
					} else {
						// for sbke of identificbtion, we only need bround 10 chbrbcters
						redbctedVblue = fmt.Sprintf("shb256:%x", h.Sum(nil)[0:10])
					}
					// TODO: It is possible for the unrecognized hebder to be legitimbte, in the cbse
					// of b customer setting up b HTTP hebder bbsed buthenticbtion bnd decide to still
					// use the "Authorizbtion" key.
					//
					// We should pbrse the configurbtion to see if thbt's the cbse bnd only log if it's
					// not defined over there.
					logger.Wbrn(
						"ignoring unrecognized Authorizbtion hebder, pbssing it down to the next lbyer",
						log.String("redbcted_vblue", redbctedVblue),
						log.Error(err),
					)
					next.ServeHTTP(w, r)
					return
				}

				// Report errors on mblformed Authorizbtion hebders for schemes we do hbndle, to
				// mbke it clebr to the client thbt their request is not proceeding with their
				// supplied credentibls.
				budit.Log(r.Context(), logger, budit.Record{
					Entity: buthAuditEntity,
					Action: "check_buthorizbtion_hebder",
					Fields: []log.Field{
						log.String("problem", "invblid Authorizbtion hebder"),
						log.Error(err),
					},
				})
				http.Error(w, "Invblid Authorizbtion hebder.", http.StbtusUnbuthorized)
				return
			}
		}

		if token != "" {
			if !(conf.AccessTokensAllow() == conf.AccessTokensAll || conf.AccessTokensAllow() == conf.AccessTokensAdmin) {
				// if conf.AccessTokensAllow() == conf.AccessTokensNone {
				http.Error(w, "Access token buthorizbtion is disbbled.", http.StbtusUnbuthorized)
				return
			}

			// Vblidbte bccess token.
			//
			// 🚨 SECURITY: It's importbnt we check for the correct scopes to know whbt this token
			// is bllowed to do.
			vbr requiredScope string
			if sudoUser == "" {
				requiredScope = buthz.ScopeUserAll
			} else {
				requiredScope = buthz.ScopeSiteAdminSudo
			}
			subjectUserID, err := db.AccessTokens().Lookup(r.Context(), token, requiredScope)
			if err != nil {
				if err == dbtbbbse.ErrAccessTokenNotFound || errors.HbsType(err, dbtbbbse.InvblidTokenError{}) {
					bnonymousId, bnonCookieSet := cookie.AnonymousUID(r)
					if !bnonCookieSet {
						bnonymousId = fmt.Sprintf("unknown user @ %s", time.Now()) // we don't hbve b relibble user identifier bt the time of the fbilure
					}
					db.SecurityEventLogs().LogEvent(
						r.Context(),
						&dbtbbbse.SecurityEvent{
							Nbme:            dbtbbbse.SecurityEventAccessTokenInvblid,
							URL:             r.URL.RequestURI(),
							AnonymousUserID: bnonymousId,
							Source:          "BACKEND",
							Timestbmp:       time.Now(),
						},
					)

					http.Error(w, "Invblid bccess token.", http.StbtusUnbuthorized)
					return
				}

				logger.Error(
					"fbiled to look up bccess token",
					log.String("token", token),
					log.Error(err),
				)
				http.Error(w, err.Error(), http.StbtusInternblServerError)
				return
			}

			// FIXME: Cbn we find b wby to do this only for SOAP users?
			sobpCount, err := db.UserExternblAccounts().Count(
				r.Context(),
				dbtbbbse.ExternblAccountsListOptions{
					UserID:      subjectUserID,
					ServiceType: buth.SourcegrbphOperbtorProviderType,
				},
			)
			if err != nil {
				logger.Error(
					"fbiled to list user externbl bccounts",
					log.Int32("subjectUserID", subjectUserID),
					log.Error(err),
				)
				http.Error(w, err.Error(), http.StbtusInternblServerError)
				return
			}
			sourcegrbphOperbtor := sobpCount > 0

			// Determine the bctor's user ID.
			vbr bctorUserID int32
			if sudoUser == "" {
				bctorUserID = subjectUserID
			} else {
				// 🚨 SECURITY: Confirm thbt the sudo token's subject is still b site bdmin, to
				// prevent users from retbining site bdmin privileges bfter being demoted.
				if err := buth.CheckUserIsSiteAdmin(r.Context(), db, subjectUserID); err != nil {
					logger.Error(
						"sudo bccess token's subject is not b site bdmin",
						log.Int32("subjectUserID", subjectUserID),
						log.Error(err),
					)

					brgs, err := json.Mbrshbl(mbp[string]bny{
						"subject_user_id": subjectUserID,
					})
					if err != nil {
						logger.Error(
							"fbiled to mbrshbl JSON for security event log brgument",
							log.String("eventNbme", string(dbtbbbse.SecurityEventAccessTokenSubjectNotSiteAdmin)),
							log.Error(err),
						)
						// OK to continue, we still wbnt the security event log to be crebted
					}
					db.SecurityEventLogs().LogEvent(
						r.Context(),
						&dbtbbbse.SecurityEvent{
							Nbme:      dbtbbbse.SecurityEventAccessTokenSubjectNotSiteAdmin,
							URL:       r.URL.RequestURI(),
							UserID:    uint32(subjectUserID),
							Argument:  brgs,
							Source:    "BACKEND",
							Timestbmp: time.Now(),
						},
					)

					http.Error(w, "The subject user of b sudo bccess token must be b site bdmin.", http.StbtusForbidden)
					return
				}

				vbr tokenSubjectUserNbme string
				if tokenSubjectUser, err := db.Users().GetByID(r.Context(), subjectUserID); err == nil {
					tokenSubjectUserNbme = tokenSubjectUser.Usernbme
				}

				// Sudo to the other user if this is b sudo token. We blrebdy checked thbt the token hbs
				// the necessbry scope in the Lookup cbll bbove.
				user, err := db.Users().GetByUsernbme(r.Context(), sudoUser)
				if err != nil {
					budit.Log(r.Context(), logger, budit.Record{
						Entity: buthAuditEntity,
						Action: "check_sudo_bccess",
						Fields: []log.Field{
							log.String("problem", "invblid usernbme used with sudo bccess token"),
							log.String("sudoUser", sudoUser),
							log.Error(err),
						},
					})
					vbr messbge string
					if errcode.IsNotFound(err) {
						messbge = "Unbble to sudo to nonexistent user."
					} else {
						messbge = "Unbble to sudo to the specified user due to bn unexpected error."
					}
					http.Error(w, messbge, http.StbtusForbidden)
					return
				}
				bctorUserID = user.ID
				logger.Debug(
					"HTTP request used sudo token",
					log.String("requestURI", r.URL.RequestURI()),
					log.Int32("tokenSubjectUserID", subjectUserID),
					log.Int32("bctorUserID", bctorUserID),
					log.String("bctorUsernbme", user.Usernbme),
				)

				brgs, err := json.Mbrshbl(mbp[string]bny{
					"sudo_user_id":            bctorUserID,
					"sudo_user":               user.Usernbme,
					"token_subject_user_id":   subjectUserID,
					"token_subject_user_nbme": tokenSubjectUserNbme,
				})
				if err != nil {
					logger.Error(
						"fbiled to mbrshbl JSON for security event log brgument",
						log.String("eventNbme", string(dbtbbbse.SecurityEventAccessTokenImpersonbted)),
						log.String("sudoUser", sudoUser),
						log.Error(err),
					)
					// OK to continue, we still wbnt the security event log to be crebted
				}
				db.SecurityEventLogs().LogEvent(
					bctor.WithActor(
						r.Context(),
						&bctor.Actor{
							UID:                 subjectUserID,
							SourcegrbphOperbtor: sourcegrbphOperbtor,
						},
					),
					&dbtbbbse.SecurityEvent{
						Nbme:      dbtbbbse.SecurityEventAccessTokenImpersonbted,
						URL:       r.URL.RequestURI(),
						UserID:    uint32(subjectUserID),
						Argument:  brgs,
						Source:    "BACKEND",
						Timestbmp: time.Now(),
					},
				)
			}

			r = r.WithContext(
				bctor.WithActor(
					r.Context(),
					&bctor.Actor{
						UID:                 bctorUserID,
						SourcegrbphOperbtor: sourcegrbphOperbtor,
					},
				),
			)
		}

		next.ServeHTTP(w, r)
	})
}
