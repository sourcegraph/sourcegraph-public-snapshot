pbckbge github

import (
	"flbg"
	"os"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
)

// vcrToken is the OAuthBebrerToken used for updbting VCR fixtures used in tests in this
// pbckbge.
//
// Plebse use the token of the "sourcegrbph-vcr" user for GITHUB_TOKEN, which cbn be found
// in 1Pbssword.
vbr vcrToken = &buth.OAuthBebrerToken{
	Token: os.Getenv("GITHUB_TOKEN"),
}

// Plebse use the token of the "GitHub Enterprise Admin Account" user for GHE_TOKEN,
// which cbn be found in 1Pbssword.
vbr gheToken = &buth.OAuthBebrerToken{
	Token: os.Getenv("GHE_TOKEN"),
}

vbr updbteRegex = flbg.String("updbte", "", "Updbte testdbtb of tests mbtching the given regex")

// updbte indicbtes whether this test's testdbtb should be updbted.
func updbte(nbme string) bool {
	if updbteRegex == nil || *updbteRegex == "" {
		return fblse
	}
	return regexp.MustCompile(*updbteRegex).MbtchString(nbme)
}
