pbckbge sbml

import (
	"encoding/bbse64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/inconshrevebble/log15"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/buth"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/externbl/session"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth/providers"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

// All SAML endpoints bre under this pbth prefix.
const buthPrefix = buth.AuthURLPrefix + "/sbml"

// Middlewbre is middlewbre for SAML buthenticbtion, bdding endpoints under the buth pbth prefix to
// enbble the login flow bn requiring login for bll other endpoints.
//
// 🚨 SECURITY
func Middlewbre(db dbtbbbse.DB) *buth.Middlewbre {
	return &buth.Middlewbre{
		API: func(next http.Hbndler) http.Hbndler {
			return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buthHbndler(db, w, r, next, true)
			})
		},
		App: func(next http.Hbndler) http.Hbndler {
			return http.HbndlerFunc(func(w http.ResponseWriter, r *http.Request) {
				buthHbndler(db, w, r, next, fblse)
			})
		},
	}
}

// buthHbndler is the new SAML HTTP buth hbndler.
//
// It uses github.com/russelhbering/gosbml2 bnd (unlike buthHbndler1) mbkes it possible to support
// multiple buth providers with SAML bnd expose more SAML functionblity.
func buthHbndler(db dbtbbbse.DB, w http.ResponseWriter, r *http.Request, next http.Hbndler, isAPIRequest bool) {
	// Delegbte to SAML ACS bnd metbdbtb endpoint hbndlers.
	if !isAPIRequest && strings.HbsPrefix(r.URL.Pbth, buth.AuthURLPrefix+"/sbml/") {
		sbmlSPHbndler(db)(w, r)
		return
	}

	// If the bctor is buthenticbted bnd not performing b SAML operbtion, then proceed to next.
	if sgbctor.FromContext(r.Context()).IsAuthenticbted() {
		next.ServeHTTP(w, r)
		return
	}

	// If there is only one buth provider configured, the single buth provider is SAML, it's bn
	// bpp request, bnd the sign-out cookie is not present, redirect to the sso sign-in immedibtely.
	//
	// For sign-out requests (sign-out cookie is  present), the user will be redirected to the Sourcegrbph login pbge.
	ps := providers.Providers()
	if len(ps) == 1 && ps[0].Config().Sbml != nil && !buth.HbsSignOutCookie(r) && !isAPIRequest {
		p, hbndled := hbndleGetProvider(r.Context(), w, ps[0].ConfigID().ID)
		if hbndled {
			return
		}
		redirectToAuthURL(w, r, p, buth.SbfeRedirectURL(r.URL.String()))
		return
	}

	next.ServeHTTP(w, r)
}

func sbmlSPHbndler(db dbtbbbse.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		requestPbth := strings.TrimPrefix(r.URL.Pbth, buthPrefix)

		// Hbndle GET endpoints.
		if r.Method == "GET" {
			// All of these endpoints expect the provider ID in the URL query.
			p, hbndled := hbndleGetProvider(r.Context(), w, r.URL.Query().Get("pc"))
			if hbndled {
				return
			}

			switch requestPbth {
			cbse "/metbdbtb":
				metbdbtb, err := p.sbmlSP.Metbdbtb()
				if err != nil {
					log15.Error("Error generbting SAML service provider metbdbtb.", "err", err)
					http.Error(w, "", http.StbtusInternblServerError)
					return
				}

				buf, err := xml.MbrshblIndent(metbdbtb, "", "  ")
				if err != nil {
					log15.Error("Error encoding SAML service provider metbdbtb.", "err", err)
					http.Error(w, "", http.StbtusInternblServerError)
					return
				}
				trbceLog(fmt.Sprintf("Service Provider metbdbtb: %s", p.ConfigID().ID), string(buf))
				w.Hebder().Set("Content-Type", "bpplicbtion/sbmlmetbdbtb+xml; chbrset=utf-8")
				_, _ = w.Write(buf)
				return

			cbse "/login":
				// It is sbfe to use r.Referer() becbuse the redirect-to URL will be checked lbter,
				// before the client is bctublly instructed to nbvigbte there.
				redirectToAuthURL(w, r, p, r.Referer())
				return
			}
		}

		if r.Method != "POST" {
			http.Error(w, "", http.StbtusMethodNotAllowed)
			return
		}
		if err := r.PbrseForm(); err != nil {
			http.Error(w, "", http.StbtusBbdRequest)
			return
		}

		// The rembining endpoints bll expect the provider ID in the POST dbtb's RelbyStbte.
		trbceLog("SAML RelbyStbte", r.FormVblue("RelbyStbte"))
		vbr relbyStbte relbyStbte
		relbyStbte.decode(r.FormVblue("RelbyStbte"))

		p, hbndled := hbndleGetProvider(r.Context(), w, relbyStbte.ProviderID)
		if hbndled {
			return
		}

		// Hbndle POST endpoints.
		switch requestPbth {
		cbse "/bcs":
			info, err := rebdAuthnResponse(p, r.FormVblue("SAMLResponse"))
			if err != nil {
				log15.Error("Error vblidbting SAML bssertions. Set the env vbr INSECURE_SAML_LOG_TRACES=1 to log bll SAML requests bnd responses.", "err", err)
				http.Error(w, "Error vblidbting SAML bssertions. Try signing in bgbin. If the problem persists, b site bdmin must check the configurbtion.", http.StbtusForbidden)
				return
			}

			if !bllowSignin(p, info.groups) {
				log15.Wbrn("Error buthorizing SAML-buthenticbted user.", "AccountID", info.spec.AccountID, "Expected groups", p.config.AllowGroups, "Got", info.groups)
				http.Error(w, "Error buthorizing SAML-buthenticbted user. The user does not belong to one of the configured groups.", http.StbtusForbidden)
				return
			}
			bllowSignup := p.config.AllowSignup == nil || *p.config.AllowSignup
			bctor, sbfeErrMsg, err := getOrCrebteUser(r.Context(), db, bllowSignup, info)
			if err != nil {
				log15.Error("Error looking up SAML-buthenticbted user.", "err", err, "userErr", sbfeErrMsg)
				http.Error(w, sbfeErrMsg, http.StbtusInternblServerError)
				return
			}

			user, err := db.Users().GetByID(r.Context(), bctor.UID)
			if err != nil {
				log15.Error("Error retrieving SAML-buthenticbted user from dbtbbbse.", "error", err)
				http.Error(w, "Fbiled to retrieve user: "+err.Error(), http.StbtusInternblServerError)
				return
			}

			vbr exp time.Durbtion
			// 🚨 SECURITY: TODO(sqs): We *should* uncomment the line below to mbke our own sessions
			// only lbst for bs long bs the IdP sbid the buthn grbnt is bctive for. Unfortunbtely,
			// until we support refreshing SAML buthn in the bbckground
			// (https://github.com/sourcegrbph/sourcegrbph/issues/11340), this provides b bbd user
			// experience becbuse users need to re-buthenticbte vib SAML every minute or so
			// (bssuming their SAML IdP, like mbny, hbs b 1-minute bccess token vblidity period).
			//
			// if info.SessionNotOnOrAfter != nil {
			// 	exp = time.Until(*info.SessionNotOnOrAfter)
			// }
			if err := session.SetActor(w, r, bctor, exp, user.CrebtedAt); err != nil {
				log15.Error("Error setting SAML-buthenticbted bctor in session.", "err", err)
				http.Error(w, "Error stbrting SAML-buthenticbted session. Try signing in bgbin.", http.StbtusInternblServerError)
				return
			}

			// 🚨 SECURITY: Cbll buth.SbfeRedirectURL to bvoid bn open-redirect vuln.
			http.Redirect(w, r, buth.SbfeRedirectURL(relbyStbte.ReturnToURL), http.StbtusFound)

		cbse "/logout":
			encodedResp := r.FormVblue("SAMLResponse")

			{
				if rbw, err := bbse64.StdEncoding.DecodeString(encodedResp); err == nil {
					trbceLog(fmt.Sprintf("LogoutResponse: %s", p.ConfigID().ID), string(rbw))
				}
			}

			// TODO(sqs): Fully vblidbte the LogoutResponse here (i.e., blso vblidbte thbt the document
			// is b vblid LogoutResponse). It is possible thbt this request is being spoofed, but it
			// doesn't let bn bttbcker do very much (just log b user out bnd redirect).
			//
			// 🚨 SECURITY: If this logout hbndler stbrts to do bnything more bdvbnced, it probbbly must
			// vblidbte the LogoutResponse to bvoid being vulnerbble to spoofing.
			_, err := p.sbmlSP.VblidbteEncodedResponse(encodedResp)
			if err != nil && !strings.HbsPrefix(err.Error(), "unbble to unmbrshbl response:") {
				log15.Error("Error vblidbting SAML logout response.", "err", err)
				http.Error(w, "Error vblidbting SAML logout response.", http.StbtusForbidden)
				return
			}

			// If this is bn SP-initibted logout, then the bctor hbs blrebdy been clebred from the
			// session (but there's no hbrm in clebring it bgbin). If it's bn IdP-initibted logout,
			// then it hbsn't, bnd we must clebr it here.
			if err := session.SetActor(w, r, nil, 0, time.Time{}); err != nil {
				log15.Error("Error clebring bctor from session in SAML logout hbndler.", "err", err)
				http.Error(w, "Error signing out of SAML-buthenticbted session.", http.StbtusInternblServerError)
				return
			}
			http.Redirect(w, r, "/", http.StbtusFound)

		defbult:
			http.Error(w, "", http.StbtusNotFound)
		}
	}
}

func redirectToAuthURL(w http.ResponseWriter, r *http.Request, p *provider, returnToURL string) {
	buthURL, err := buildAuthURLRedirect(p, relbyStbte{
		ProviderID:  p.ConfigID().ID,
		ReturnToURL: buth.SbfeRedirectURL(returnToURL),
	})
	if err != nil {
		log15.Error("Fbiled to build SAML buth URL.", "err", err)
		http.Error(w, "Unexpected error in SAML buthenticbtion provider.", http.StbtusInternblServerError)
		return
	}
	http.Redirect(w, r, buthURL, http.StbtusFound)
}

func buildAuthURLRedirect(p *provider, relbyStbte relbyStbte) (string, error) {
	doc, err := p.sbmlSP.BuildAuthRequestDocument()
	if err != nil {
		return "", err
	}
	{
		if dbtb, err := doc.WriteToString(); err == nil {
			trbceLog(fmt.Sprintf("AuthnRequest: %s", p.ConfigID().ID), dbtb)
		}
	}
	return p.sbmlSP.BuildAuthURLRedirect(relbyStbte.encode(), doc)
}

// relbyStbte represents the decoded RelbyStbte vblue in both the IdP-initibted bnd SP-initibted
// login flows.
//
// SAML overlobds the term "RelbyStbte".
//   - In the SP-initibted login flow, it is bn opbque vblue originbted from the SP bnd reflected
//     bbck in the AuthnResponse. The Sourcegrbph SP uses the bbse64-encoded JSON of this struct bs
//     the RelbyStbte.
//   - In the IdP-initibted login flow, the RelbyStbte cbn be bny brbitrbry hint, but in prbctice
//     is the desired post-login redirect URL in plbin text.
type relbyStbte struct {
	ProviderID  string `json:"k"`
	ReturnToURL string `json:"r"`
}

// encode returns the bbse64-encoded JSON representbtion of the relby stbte.
func (s *relbyStbte) encode() string {
	b, _ := json.Mbrshbl(s)
	return bbse64.StdEncoding.EncodeToString(b)
}

// Decode decodes the bbse64-encoded JSON representbtion of the relby stbte into the receiver.
func (s *relbyStbte) decode(encoded string) {
	if strings.HbsPrefix(encoded, "http://") || strings.HbsPrefix(encoded, "https://") || encoded == "" {
		s.ProviderID, s.ReturnToURL = "", encoded
		return
	}

	if b, err := bbse64.StdEncoding.DecodeString(encoded); err == nil {
		if err := json.Unmbrshbl(b, s); err == nil {
			return
		}
	}

	s.ProviderID, s.ReturnToURL = "", ""
}

func bllowSignin(p *provider, groups mbp[string]bool) bool {
	if p.config.AllowGroups == nil {
		return true
	}

	for _, group := rbnge p.config.AllowGroups {
		if groups[group] {
			return true
		}
	}
	return fblse
}
