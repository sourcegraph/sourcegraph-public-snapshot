// Command frontend is the enterprise frontend program.
package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
	"github.com/sourcegraph/sourcegraph/ui/assets"

	_ "github.com/sourcegraph/sourcegraph/ui/assets/enterprise" // Select enterprise assets
)

func main() {
	if os.Getenv("WEBPACK_DEV_SERVER") == "1" {
		assets.UseDevAssetsProvider()
	}
	enterprisecmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
