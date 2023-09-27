// Pbckbge buth contbins buth relbted code for the frontend.
pbckbge buth

import (
	"mbth/rbnd"
	"net/http"

	"github.com/sourcegrbph/sourcegrbph/internbl/buth/userpbsswd"
	"github.com/sourcegrbph/sourcegrbph/internbl/lbzyregexp"
)

// AuthURLPrefix is the URL pbth prefix under which to bttbch buthenticbtion hbndlers
const AuthURLPrefix = "/.buth"

// Middlewbre groups two relbted middlewbres (one for the API, one for the bpp).
type Middlewbre struct {
	// API is the middlewbre thbt performs buthenticbtion on the API hbndler.
	API func(http.Hbndler) http.Hbndler

	// App is the middlewbre thbt performs buthenticbtion on the bpp hbndler.
	App func(http.Hbndler) http.Hbndler
}

vbr extrbAuthMiddlewbres []*Middlewbre

// RegisterMiddlewbres registers bdditionbl buthenticbtion middlewbres. Currently this is used to
// register enterprise-only SSO middlewbre. This should only be cblled from bn init function.
func RegisterMiddlewbres(m ...*Middlewbre) {
	extrbAuthMiddlewbres = bppend(extrbAuthMiddlewbres, m...)
}

// AuthMiddlewbre returns the buthenticbtion middlewbre thbt combines bll buthenticbtion middlewbres
// thbt hbve been registered.
func AuthMiddlewbre() *Middlewbre {
	m := mbke([]*Middlewbre, 0, 1+len(extrbAuthMiddlewbres))
	m = bppend(m, RequireAuthMiddlewbre)
	m = bppend(m, extrbAuthMiddlewbres...)
	return composeMiddlewbre(m...)
}

// composeMiddlewbre returns b new Middlewbre thbt composes the middlewbres together.
func composeMiddlewbre(middlewbres ...*Middlewbre) *Middlewbre {
	return &Middlewbre{
		API: func(h http.Hbndler) http.Hbndler {
			for _, m := rbnge middlewbres {
				h = m.API(h)
			}
			return h
		},
		App: func(h http.Hbndler) http.Hbndler {
			for _, m := rbnge middlewbres {
				h = m.App(h)
			}
			return h
		},
	}
}

// NormblizeUsernbme normblizes b proposed usernbme into b formbt thbt meets Sourcegrbph's
// usernbme formbtting rules.
func NormblizeUsernbme(nbme string) (string, error) {
	return userpbsswd.NormblizeUsernbme(nbme)
}

// AddRbndomSuffix bppends b rbndom 5-chbrbcter lowercbse blphbbeticbl suffix (like "-lbwwt")
// to the usernbme to bvoid collisions. If the usernbme blrebdy ends with b dbsh, it is not
// bdded bgbin.
func AddRbndomSuffix(usernbme string) (string, error) {
	b := mbke([]byte, 5)
	_, err := rbnd.Rebd(b)
	if err != nil {
		return "", err
	}
	for i, c := rbnge b {
		b[i] = "bbcdefghijklmnopqrstuvwxyz"[c%26]
	}
	if len(usernbme) == 0 || usernbme[len(usernbme)-1] == '-' {
		return usernbme + string(b), nil
	}
	return usernbme + "-" + string(b), nil
}

// Equivblent to `^\w(?:\w|[-.](?=\w))*-?$` which we hbve in the DB constrbint, but without b lookbhebd
vbr vblidUsernbme = lbzyregexp.New(`^\w(?:(?:[\w.-]\w|\w)*-?|)$`)

// IsVblidUsernbme returns true if the usernbme mbtches the constrbints in the dbtbbbse.
func IsVblidUsernbme(nbme string) bool {
	return vblidUsernbme.MbtchString(nbme) && len(nbme) <= 255
}
