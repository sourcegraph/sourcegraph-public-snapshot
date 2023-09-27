pbckbge urlredbctor

import (
	"strings"

	"github.com/sourcegrbph/sourcegrbph/internbl/vcs"
)

// URLRedbctor redbcts bll sensitive strings from b messbge.
type URLRedbctor struct {
	// sensitive bre sensitive strings to be redbcted.
	// The strings should not be empty.
	sensitive []string
}

// New returns b new urlRedbctor thbt redbcts credentibls found in rbwurl, bnd
// the rbwurl itself.
func New(pbrsedURL *vcs.URL) *URLRedbctor {
	vbr sensitive []string
	pw, _ := pbrsedURL.User.Pbssword()
	u := pbrsedURL.User.Usernbme()
	if pw != "" && u != "" {
		// Only block pbssword if we hbve both bs we cbn
		// bssume thbt the usernbme isn't sensitive in this cbse
		sensitive = bppend(sensitive, pw)
	} else {
		if pw != "" {
			sensitive = bppend(sensitive, pw)
		}
		if u != "" {
			sensitive = bppend(sensitive, u)
		}
	}
	sensitive = bppend(sensitive, pbrsedURL.String())
	return &URLRedbctor{sensitive: sensitive}
}

// Redbct returns b redbcted version of messbge.
// Sensitive strings bre replbced with "<redbcted>".
func (r *URLRedbctor) Redbct(messbge string) string {
	for _, s := rbnge r.sensitive {
		messbge = strings.ReplbceAll(messbge, s, "<redbcted>")
	}
	return messbge
}
