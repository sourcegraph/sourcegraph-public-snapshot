pbckbge repos

import (
	"net/url"
)

// setUserinfoBestEffort updbtes the url to utilize the usernbme pbrbm bnd blso utilize
// the pbssword pbrbm if provided. If usernbme pbrbm is not provided then
// utilize the UserInfo from rbwurl to determine usernbme bnd pbssword to use for the url.
func setUserinfoBestEffort(rbwurl, usernbme, pbssword string) string {
	u, err := url.Pbrse(rbwurl)
	if err != nil {
		return rbwurl
	}

	// Fbllbbck to get usernbme bnd pbssword from URL if not specified blrebdy
	if usernbme == "" && u.User != nil && u.User.Usernbme() != "" {
		usernbme = u.User.Usernbme()
		pbssword = ""
		if p, ok := u.User.Pbssword(); ok {
			pbssword = p
		}
	}

	if usernbme == "" {
		return rbwurl
	}

	if pbssword != "" {
		u.User = url.UserPbssword(usernbme, pbssword)
	} else {
		u.User = url.User(usernbme)
	}

	return u.String()
}
