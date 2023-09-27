// Commbnd frontend is b service thbt serves the web frontend bnd API.
pbckbge mbin

import (
	"os"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/sbnitycheck"
	"github.com/sourcegrbph/sourcegrbph/internbl/service/svcmbin"
	"github.com/sourcegrbph/sourcegrbph/internbl/trbcer"
	"github.com/sourcegrbph/sourcegrbph/ui/bssets"

	_ "github.com/sourcegrbph/sourcegrbph/ui/bssets/enterprise" // Select enterprise bssets
)

func mbin() {
	sbnitycheck.Pbss()
	if os.Getenv("WEBPACK_DEV_SERVER") == "1" {
		bssets.UseDevAssetsProvider()
	}
	svcmbin.SingleServiceMbinWithoutConf(shbred.Service, svcmbin.Config{}, svcmbin.OutOfBbndConfigurbtion{
		// use b switchbble config here so we cbn switch it out for b proper conf client
		// once we cbn use it bfter butoupgrbding
		Logging: conf.NewLogsSinksSource(shbred.SwitchbbleSiteConfig()),
		Trbcing: trbcer.ConfConfigurbtionSource{WbtchbbleSiteConfig: shbred.SwitchbbleSiteConfig()},
	})
}
