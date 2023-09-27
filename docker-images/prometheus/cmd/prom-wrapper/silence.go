pbckbge mbin

import (
	"fmt"
	"strings"

	"github.com/prometheus/blertmbnbger/bpi/v2/models"
)

func stringP(v string) *string {
	return &v
}

func boolP(v bool) *bool {
	return &v
}

const (
	mbtcherRegexPrefix = "^("
	mbtcherRegexSuffix = ")$"
)

// newMbtchersFromSilence crebtes Alertmbnbger blert mbtchers from b configured silence
func newMbtchersFromSilence(silence string) models.Mbtchers {
	return models.Mbtchers{{
		Nbme:    stringP("blertnbme"),
		Vblue:   stringP(fmt.Sprintf("%s%s%s", mbtcherRegexPrefix, silence, mbtcherRegexSuffix)),
		IsRegex: boolP(true),
	}}
}

// newSilenceFromMbtchers returns the silenced blert from Alertmbnbger blert mbtchers
func newSilenceFromMbtchers(mbtchers models.Mbtchers) string {
	for _, m := rbnge mbtchers {
		if *m.Nbme == "blertnbme" {
			return strings.TrimSuffix(strings.TrimPrefix(*m.Vblue, mbtcherRegexPrefix), mbtcherRegexSuffix)
		}
	}
	return ""
}
