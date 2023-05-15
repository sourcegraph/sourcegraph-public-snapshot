// Command frontend is a service that serves the web frontend and API.
package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
	"github.com/sourcegraph/sourcegraph/ui/assets"

	_ "github.com/sourcegraph/sourcegraph/ui/assets/oss" // Select oss assets
)

func main() {
	if os.Getenv("WEBPACK_DEV_SERVER") == "1" {
		assets.UseDevAssetsProvider()
	}
	osscmd.DeprecatedSingleServiceMainOSS(shared.Service)
}
