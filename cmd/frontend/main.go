// Command frontend is a service that serves the web frontend and API.
package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
	"github.com/sourcegraph/sourcegraph/internal/tracer"
	"github.com/sourcegraph/sourcegraph/ui/assets"

	_ "github.com/sourcegraph/sourcegraph/client/web/dist" // use assets
)

func main() {
	sanitycheck.Pass()
	if os.Getenv("WEB_BUILDER_DEV_SERVER") == "1" {
		assets.UseDevAssetsProvider()
	}
	svcmain.SingleServiceMainWithoutConf(shared.Service, svcmain.Config{}, svcmain.OutOfBandConfiguration{
		// use a switchable config here so we can switch it out for a proper conf client
		// once we can use it after autoupgrading
		Logging: conf.NewLogsSinksSource(shared.SwitchableSiteConfig()),
		Tracing: tracer.ConfConfigurationSource{WatchableSiteConfig: shared.SwitchableSiteConfig()},
	})
}
