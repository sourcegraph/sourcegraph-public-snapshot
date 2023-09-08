// Command frontend is a service that serves the web frontend and API.
package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph/osscmd"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/ui/assets"

	_ "github.com/sourcegraph/sourcegraph/ui/assets/oss" // Select oss assets
)

func main() {
	sanitycheck.Pass()

	if os.Getenv("WEBPACK_DEV_SERVER") == "1" {
		assets.UseDevAssetsProvider()
	}
	osscmd.SingleServiceMainOSS(shared.Service)
}
