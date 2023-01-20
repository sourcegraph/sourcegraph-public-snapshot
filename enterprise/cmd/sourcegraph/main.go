package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
	"github.com/sourcegraph/sourcegraph/internal/service"

	githubproxy_shared "github.com/sourcegraph/sourcegraph/cmd/github-proxy/shared"
	searcher_shared "github.com/sourcegraph/sourcegraph/cmd/searcher/shared"
	executor_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/executor/shared"
	frontend_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/shared"
	gitserver_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/gitserver/shared"
	precise_code_intel_worker_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/shared"
	repoupdater_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/shared"
	symbols_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols/shared"
	worker_shared "github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared"
)

// services is a list of services to run in the enterprise build.
var services = []service.Service{
	frontend_shared.Service,
	gitserver_shared.Service,
	repoupdater_shared.Service,
	searcher_shared.Service,
	symbols_shared.Service,
	worker_shared.Service,
	githubproxy_shared.Service,
	precise_code_intel_worker_shared.Service,
	executor_shared.Service,
}

func main() {
	enterprisecmd.MainEnterprise(services)
}
