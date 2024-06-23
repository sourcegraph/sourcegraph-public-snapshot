package main

import (
	"fmt"
	"os"

	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"

	frontend_shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	gitserver_shared "github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	repoupdater_shared "github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	searcher_shared "github.com/sourcegraph/sourcegraph/cmd/searcher/shared"
	worker_shared "github.com/sourcegraph/sourcegraph/cmd/worker/shared"

	_ "github.com/sourcegraph/sourcegraph/client/web/dist" // use assets
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

// services is a list of services to run (in addition to `frontend`).
var services = []service.Service{
	gitserver_shared.Service,
	repoupdater_shared.Service,
	searcher_shared.Service,
	worker_shared.Service,
}

func main() {
	sanitycheck.Pass()
	if os.Getenv("WEB_BUILDER_DEV_SERVER") == "1" {
		assets.UseDevAssetsProvider()
	}
	fmt.Fprintln(os.Stderr, "✱ Sourcegraph (single-program dev)")
	svcmain.FrontendMain(frontend_shared.Service, services)
}
