package main

import (
	"os"

	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/service/servegit"

	blobstore_shared "github.com/sourcegraph/sourcegraph/cmd/blobstore/shared"
	githubproxy_shared "github.com/sourcegraph/sourcegraph/cmd/github-proxy/shared"
	searcher_shared "github.com/sourcegraph/sourcegraph/cmd/searcher/shared"
	embeddings_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/embeddings/shared"
	executor_singlebinary "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/singlebinary"
	frontend_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/shared"
	gitserver_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/gitserver/shared"
	precise_code_intel_worker_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/shared"
	repoupdater_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/shared"
	symbols_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols/shared"
	worker_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared"

	"github.com/sourcegraph/sourcegraph/ui/assets"
	_ "github.com/sourcegraph/sourcegraph/ui/assets/enterprise" // Select enterprise assets
)

// services is a list of services to run in the enterprise build.
var services = []service.Service{
	frontend_shared.Service,
	gitserver_shared.Service,
	repoupdater_shared.Service,
	searcher_shared.Service,
	blobstore_shared.Service,
	symbols_shared.Service,
	worker_shared.Service,
	githubproxy_shared.Service,
	precise_code_intel_worker_shared.Service,
	executor_singlebinary.Service,
	servegit.Service,
	embeddings_shared.Service,
}

func main() {
	if os.Getenv("WEBPACK_DEV_SERVER") == "1" {
		assets.UseDevAssetsProvider()
	}
	enterprisecmd.MainEnterprise(services, os.Args)
}
