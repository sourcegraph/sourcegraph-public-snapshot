// Command sourcegraph-oss is a single program that runs all of Sourcegraph (OSS variant).
package main

import (
	frontend_shared "github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	githubproxy_shared "github.com/sourcegraph/sourcegraph/cmd/github-proxy/shared"
	gitserver_shared "github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	repoupdater_shared "github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	searcher_shared "github.com/sourcegraph/sourcegraph/cmd/searcher/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
	symbols_shared "github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
	worker_shared "github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/internal/service"
)

// services is a list of services to run in the OSS build.
var services = []service.Service{
	frontend_shared.Service,
	gitserver_shared.Service,
	repoupdater_shared.Service,
	searcher_shared.Service,
	symbols_shared.Service,
	worker_shared.Service,
	githubproxy_shared.Service,
}

func main() {
	osscmd.MainOSS(services)
}
