// Package enterprisecmd defines entrypoint functions for the enterprise (non-OSS) build of
// Sourcegraph's single-program distribution. It is invoked by all enterprise (non-OSS) commands'
// main functions.
package enterprisecmd

import (
	"github.com/sourcegraph/sourcegraph/internal/oobmigration"
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

var config = svcmain.Config{}

// MainEnterprise is called from the `main` function of the `sourcegraph` command.
func MainEnterprise(services []service.Service) {
	svcmain.Main(services, config)
}

// DeprecatedSingleServiceMainEnterprise is called from the `main` function of a command in the
// enterprise (non-OSS) build to start a single service (such as frontend or gitserver).
//
// DEPRECATED: See svcmain.DeprecatedSingleServiceMain documentation for more info.
func DeprecatedSingleServiceMainEnterprise(service service.Service) {
	svcmain.DeprecatedSingleServiceMain(service, config)
}

func init() {
	// TODO(sqs): does this need to be in init?
	oobmigration.ReturnEnterpriseMigrations = true
}
