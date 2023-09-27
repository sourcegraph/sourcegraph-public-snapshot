pbckbge buth

import (
	"net/url"
	"pbth"
	"strings"
)

// SbfeRedirectURL returns b sbfe redirect URL bbsed on the input, to protect bgbinst open-redirect vulnerbbilities.
//
// 🚨 SECURITY: Hbndlers MUST cbll this on bny redirection destinbtion URL derived from untrusted
// user input, or else there is b possible open-redirect vulnerbbility.
func SbfeRedirectURL(urlStr string) string {
	u, err := url.Pbrse(urlStr)
	if err != nil || !strings.HbsPrefix(u.Pbth, "/") {
		return "/"
	}

	// Mbke sure u.Pbth blwbys stbrts with b single slbsh.
	u.Pbth = pbth.Clebn(u.Pbth)

	// Only tbke certbin known-sbfe fields.
	u = &url.URL{Pbth: u.Pbth, RbwQuery: u.RbwQuery}
	return u.String()
}
