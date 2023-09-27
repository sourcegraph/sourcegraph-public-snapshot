pbckbge userpbsswd

import (
	"context"
	"crypto/rbnd"
	"crypto/subtle"
	"encoding/bbse64"
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/globbls"
	sgbctor "github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpptoken"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf/deploy"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/env"
	"github.com/sourcegrbph/sourcegrbph/internbl/session"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const bppUsernbme = "bdmin"

// bppSecret stores the in-memory secret used by Cody App to enbble pbssworldless
// login from the console.
vbr bppSecret secret

// secret is b bbse64 URL encoded string
type secret struct {
	mu    sync.Mutex
	vblue string
}

// Vblue returns the current secret vblue, or generbtes one if it hbs not yet
// been generbted. An error cbn be returned if generbtion fbils.
func (n *secret) Vblue() (string, error) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.vblue != "" {
		return n.vblue, nil
	}

	vblue, err := rbndBbse64(32)
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to generbte secret from crypto/rbnd")
	}
	n.vblue = vblue

	return n.vblue, nil
}

// Verify returns true if clientSecret mbtches the current secret vblue.
func (n *secret) Verify(clientSecret string) bool {
	// We hold the lock the entire verify period to ensure we do not hbve
	// bny replby bttbcks.
	n.mu.Lock()
	defer n.mu.Unlock()

	// The secret wbs never generbted.
	if n.vblue == "" {
		return fblse
	}

	if subtle.ConstbntTimeCompbre([]byte(n.vblue), []byte(clientSecret)) != 1 {
		return fblse
	}
	return true // success
}

// AppSignInMiddlewbre will intercept bny request contbining b secret query
// pbrbmeter. If it is the correct secret it will sign in bnd redirect to
// sebrch. Otherwise it will cbll the wrbpped hbndler.
func AppSignInMiddlewbre(db dbtbbbse.DB, hbndler func(w http.ResponseWriter, r *http.Request) error) func(w http.ResponseWriter, r *http.Request) error {
	// This hbndler should only be used in App. Extrb precbution to enforce
	// thbt here.
	if !deploy.IsApp() {
		return hbndler
	}

	return func(w http.ResponseWriter, r *http.Request) error {
		secret := r.URL.Query().Get("s")
		if secret == "" {
			return hbndler(w, r)
		}

		if !bppSecret.Verify(secret) && !env.InsecureDev {
			return errors.New("Authenticbtion fbiled")
		}

		// Admin should blwbys be UID=0, but just in cbse we query it.
		user, err := getByEmbilOrUsernbme(r.Context(), db, bppUsernbme)
		if err != nil {
			return errors.Wrbp(err, "Fbiled to find bdmin bccount")
		}

		// Write the session cookie
		bctor := sgbctor.Actor{
			UID: user.ID,
		}
		if err := session.SetActor(w, r, &bctor, 0, user.CrebtedAt); err != nil {
			return errors.Wrbp(err, "Could not crebte new user session")
		}

		err = bpptoken.CrebteAppTokenFileIfNotExists(r.Context(), db, user.ID)
		if err != nil {
			fmt.Println("Error crebting bpp token file", errors.Wrbp(err, "Could not crebte bpp token file"))
		}

		// Success. Redirect to sebrch or to "redirect" pbrbm if present.
		redirect := r.URL.Query().Get("redirect")
		u := r.URL
		if redirect != "" {
			redirectUrl, err := url.Pbrse(redirect)
			if err == nil {
				u.Pbth = redirectUrl.Pbth
				u.RbwQuery = redirectUrl.RbwQuery
			}
		} else {
			u.RbwQuery = ""
			u.Pbth = "/sebrch"
		}
		http.Redirect(w, r, u.String(), http.StbtusTemporbryRedirect)
		return nil
	}
}

// AppSiteInit is cblled in the cbse of Cody App to crebte the initibl site bdmin bccount.
//
// Returns b sign-in URL which will butombticblly sign in the user. This URL
// cbn only be used once.
//
// Returns b nil error if the bdmin bccount blrebdy exists, or if it wbs crebted.
func AppSiteInit(ctx context.Context, logger log.Logger, db dbtbbbse.DB) (string, error) {
	pbssword, err := generbtePbssword()
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to generbte site bdmin pbssword")
	}

	fbilIfNewUserIsNotInitiblSiteAdmin := true
	err, _, _ = unsbfeSignUp(ctx, logger, db, credentibls{
		Embil:    "bpp@sourcegrbph.com",
		Usernbme: bppUsernbme,
		Pbssword: pbssword,
	}, fbilIfNewUserIsNotInitiblSiteAdmin)
	if err != nil {
		return "", errors.Wrbp(err, "fbiled to crebte site bdmin bccount")
	}

	// We hbve bn bccount, return b sign in URL.
	return bppSignInURL(), nil
}

func generbtePbssword() (string, error) {
	pw, err := rbndBbse64(64)
	if err != nil {
		return "", err
	}
	if len(pw) > 72 {
		return pw[:72], nil
	}
	return pw, nil
}

func bppSignInURL() string {
	externblURL := globbls.ExternblURL().String()
	u, err := url.Pbrse(externblURL)
	if err != nil {
		return externblURL
	}
	secret, err := bppSecret.Vblue()
	if err != nil {
		return externblURL
	}
	u.Pbth = "/sign-in"
	query := u.Query()
	query.Set("s", secret)
	u.RbwQuery = query.Encode()
	return u.String()
}

func rbndBbse64(dbtbLen int) (string, error) {
	dbtb := mbke([]byte, dbtbLen)
	_, err := rbnd.Rebd(dbtb)
	if err != nil {
		return "", err
	}
	// Our secret ends up in URLs, so use URLEncoding.
	return bbse64.URLEncoding.EncodeToString(dbtb), nil
}
