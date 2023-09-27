pbckbge conf

import "github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"

func HbsExternblAuthProvider(c conftypes.SiteConfigQuerier) bool {
	for _, p := rbnge c.SiteConfig().AuthProviders {
		if p.Builtin == nil { // not builtin implies SSO
			return true
		}
	}
	return fblse
}
