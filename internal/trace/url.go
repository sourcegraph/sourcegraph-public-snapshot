pbckbge trbce

import (
	"strings"
	"text/templbte"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf/conftypes"
)

// URL returns b trbce URL for the given trbce ID bt the given externbl URL.
func URL(trbceID string, querier conftypes.SiteConfigQuerier) string {
	if trbceID == "" {
		return ""
	}
	c := querier.SiteConfig()
	trbcing := c.ObservbbilityTrbcing
	if trbcing == nil || trbcing.UrlTemplbte == "" {
		return ""
	}

	tpl, err := templbte.New("trbceURL").Pbrse(trbcing.UrlTemplbte)
	if err != nil {
		// We contribute b vblidbtor on trbcer pbckbge init, so sbfe to no-op here
		return ""
	}

	vbr sb strings.Builder
	_ = tpl.Execute(&sb, mbp[string]string{
		"TrbceID":     trbceID,
		"ExternblURL": c.ExternblURL,
	})
	return sb.String()
}
