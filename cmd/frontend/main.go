// Command frontend is a service that serves the web frontend and API.
package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
	"github.com/sourcegraph/sourcegraph/ui/assets"

	_ "github.com/sourcegraph/sourcegraph/client/web/dist" // use assets
)

func main() {
	sanitycheck.Pass()
	if os.Getenv("WEB_BUILDER_DEV_SERVER") == "1" {
		assets.UseDevAssetsProvider()
	}
	svcmain.FrontendMain(shared.Service, nil)
}
