pbckbge sourcegrbphoperbtor

import (
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/cloud"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestVblidbteConfig(t *testing.T) {
	cloud.MockSiteConfig(
		t,
		&cloud.SchembSiteConfig{
			AuthProviders: &cloud.SchembAuthProviders{
				SourcegrbphOperbtor: &cloud.SchembAuthProviderSourcegrbphOperbtor{
					Issuer: "https://exbmple.com/blice",
				},
			},
		},
	)

	conf.TestVblidbtor(
		t,
		conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{},
		},
		vblidbteConfig,
		conf.NewSiteProblems("Sourcegrbph Operbtor buthenticbtion provider requires `externblURL` to be set to the externbl URL of your site (exbmple: https://sourcegrbph.exbmple.com)"),
	)
}
