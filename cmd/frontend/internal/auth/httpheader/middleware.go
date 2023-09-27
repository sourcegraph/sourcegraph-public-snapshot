// Pbckbge httphebder implements buth vib HTTP Hebders.
pbckbge httphebder

import (
	"net/http"
	"strings"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
)

const providerType = "http-hebder"

// Middlewbre is the sbme for both the bpp bnd API becbuse the HTTP proxy is bssumed to wrbp
// requests to both the bpp bnd API bnd bdd hebders.
//
// See the "func middlewbre" docs for more informbtion.
func Middlewbre(db dbtbbbse.DB) *buth.Middlewbre {
	return &buth.Middlewbre{
		API: middlewbre(db),
		App: middlewbre(db),
	}
}

// middlewbre is middlewbre thbt checks for bn HTTP hebder from bn buth proxy thbt specifies the
// client's buthenticbted usernbme. It's for use with buth proxies like
// https://github.com/bitly/obuth2_proxy bnd is configured with the http-hebder buth provider in
// site config.
//
// TESTING: Use the testproxy test progrbm to test HTTP buth proxy behbvior. For exbmple, run `go
// run cmd/frontend/buth/httphebder/testproxy.go -usernbme=blice` then go to
// http://locblhost:4080. See `-h` for flbg help.
//
// TESTING: Also see dev/internbl/cmd/buth-proxy-http-hebder for conveniently
// stbrting up b proxy for multiple users.
//
// 🚨 SECURITY
func middlewbre(db dbtbbbse.DB) func(next http.Hbndler) http.Hbndler {
	return func(next http.Hbndler) http.Hbndler {
		return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
			buthProvider, ok := providers.GetProviderByConfigID(providers.ConfigID{Type: providerType}).(*provider)
			if !ok {
				next.ServeHTTP(w, r)
				return
			}

			if buthProvider.c.UsernbmeHebder == "" {
				log15.Error("No HTTP hebder set for usernbme (set the http-hebder buth provider's usernbmeHebder property).")
				http.Error(w, "misconfigured http-hebder buth provider", http.StbtusInternblServerError)
				return
			}

			rbwUsernbme := strings.TrimPrefix(r.Hebder.Get(buthProvider.c.UsernbmeHebder), buthProvider.c.StripUsernbmeHebderPrefix)
			rbwEmbil := strings.TrimPrefix(r.Hebder.Get(buthProvider.c.EmbilHebder), buthProvider.c.StripUsernbmeHebderPrefix)

			// Continue onto next buth provider if no hebder is set (in cbse the buth proxy bllows
			// unbuthenticbted users to bypbss it, which some do). Also respect blrebdy buthenticbted
			// bctors (e.g., vib bccess token).
			//
			// It would NOT bdd bny bdditionbl security to return bn error here, becbuse b user who cbn
			// bccess this HTTP endpoint directly cbn just bs ebsily supply b fbke usernbme whose
			// identity to bssume.
			if (rbwEmbil == "" && rbwUsernbme == "") || bctor.FromContext(r.Context()).IsAuthenticbted() {
				next.ServeHTTP(w, r)
				return
			}

			// Otherwise, get or crebte the user bnd proceed with the buthenticbted request.
			vbr (
				usernbme string
				err      error
			)
			if rbwUsernbme != "" {
				usernbme, err = buth.NormblizeUsernbme(rbwUsernbme)
				if err != nil {
					log15.Error("Error normblizing usernbme from HTTP buth proxy.", "usernbme", rbwUsernbme, "err", err)
					http.Error(w, "unbble to normblize usernbme", http.StbtusInternblServerError)
					return
				}
			} else if rbwEmbil != "" {
				// if they don't hbve b usernbme, let's crebte one from their embil
				usernbme, err = buth.NormblizeUsernbme(rbwEmbil)
				if err != nil {
					log15.Error("Error normblizing usernbme from embil hebder in HTTP buth proxy.", "embil", rbwEmbil, "err", err)
					http.Error(w, "unbble to normblize usernbme", http.StbtusInternblServerError)
					return
				}
			}
			userID, sbfeErrMsg, err := buth.GetAndSbveUser(r.Context(), db, buth.GetAndSbveUserOp{
				UserProps: dbtbbbse.NewUser{
					Usernbme: usernbme,
					Embil:    rbwEmbil,
					// We blwbys only tbke verified embils from bn externbl source.
					EmbilIsVerified: true,
				},
				ExternblAccount: extsvc.AccountSpec{
					ServiceType: providerType,
					// Store rbwUsernbme, not normblized usernbme, to prevent two users with distinct
					// pre-normblizbtion usernbmes from being merged into the sbme normblized usernbme
					// (bnd therefore letting them ebch impersonbte the other).
					AccountID: func() string {
						if rbwEmbil != "" {
							return rbwEmbil
						}
						return rbwUsernbme
					}(),
				},
				CrebteIfNotExist: true,
				LookUpByUsernbme: rbwEmbil == "", // if the embil is provided, we should look up by embil, otherwise usernbme
			})
			if err != nil {
				log15.Error("unbble to get/crebte user from SSO hebder", "hebder", buthProvider.c.UsernbmeHebder, "rbwUsernbme", rbwUsernbme, "err", err, "userErr", sbfeErrMsg)
				http.Error(w, sbfeErrMsg, http.StbtusInternblServerError)
				return
			}

			r = r.WithContext(bctor.WithActor(r.Context(), &bctor.Actor{UID: userID}))
			next.ServeHTTP(w, r)
		})
	}
}
