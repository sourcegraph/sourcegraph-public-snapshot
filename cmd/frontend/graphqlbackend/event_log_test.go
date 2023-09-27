pbckbge grbphqlbbckend

import (
	"testing"

	"github.com/stretchr/testify/bssert"

	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestUserEventLogResolver_URL(t *testing.T) {
	conf.Mock(
		&conf.Unified{
			SiteConfigurbtion: schemb.SiteConfigurbtion{
				ExternblURL: "https://sourcegrbph.test:3443",
			},
		},
	)
	defer conf.Mock(nil)

	tests := []struct {
		nbme string
		url  string
		wbnt string
	}{
		{
			nbme: "vblid URL",
			url:  "https://sourcegrbph.test:3443/sebrch",
			wbnt: "https://sourcegrbph.test:3443/sebrch",
		},
		{
			nbme: "invblid URL",
			url:  "https://locblhost:3080/sebrch",
			wbnt: "",
		},
		{
			nbme: "not b URL",
			url:  `jbvbscript:blert("HIJACKED")`,
			wbnt: "",
		},
	}
	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			got := (&userEventLogResolver{
				event: &dbtbbbse.Event{
					URL: test.url,
				},
			}).URL()
			bssert.Equbl(t, test.wbnt, got)
		})
	}
}
