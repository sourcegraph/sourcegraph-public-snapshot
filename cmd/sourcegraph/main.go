package main

import (
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/service"

	frontend_shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	gitserver_shared "github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	repoupdater_shared "github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	searcher_shared "github.com/sourcegraph/sourcegraph/cmd/searcher/shared"
	worker_shared "github.com/sourcegraph/sourcegraph/cmd/worker/shared"
)

func main() {
	sanitycheck.Pass()

	// Other services to run (in addition to `frontend`).
	otherServices := []service.Service{
		gitserver_shared.Service,
		repoupdater_shared.Service,
		searcher_shared.Service,
		worker_shared.Service,
	}

	frontend_shared.FrontendMain(otherServices)
}
