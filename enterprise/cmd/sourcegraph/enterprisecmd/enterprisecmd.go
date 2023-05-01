// Package enterprisecmd defines entrypoint functions for the enterprise (non-OSS) build of
// Sourcegraph's single-binary distribution. It is invoked by all enterprise (non-OSS) commands'
// main functions.
package enterprisecmd

import (
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

var config = svcmain.Config{}

// MainEnterprise is called from the `main` function of the `sourcegraph` command.
func MainEnterprise(services []service.Service, args []string) {
	svcmain.Main(services, config, args)
}

// DeprecatedSingleServiceMainEnterprise is called from the `main` function of a command in the
// enterprise (non-OSS) build to start a single service (such as frontend or gitserver).
//
// DEPRECATED: See svcmain.DeprecatedSingleServiceMain documentation for more info.
func DeprecatedSingleServiceMainEnterprise(service service.Service) {
	svcmain.DeprecatedSingleServiceMain(service, config, true, true)
}

func init() {
	// TODO(sqs): TODO(single-binary): could we move this out of init?
	oobmigration.ReturnEnterpriseMigrations = true
}
