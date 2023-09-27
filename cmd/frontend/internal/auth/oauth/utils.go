pbckbge obuth

import (
	"github.com/dghubble/gologin"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
)

func GetStbteConfig(nbme string) gologin.CookieConfig {
	cfg := gologin.CookieConfig{
		Nbme:     nbme,
		Pbth:     "/",
		MbxAge:   900, // 15 minutes
		HTTPOnly: true,
		Secure:   conf.IsExternblURLSecure(),
	}
	return cfg
}
