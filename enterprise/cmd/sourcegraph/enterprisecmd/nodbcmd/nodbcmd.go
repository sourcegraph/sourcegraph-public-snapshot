// Package nodbcmd is identical to enterprisecmd, except that it does not import the dbconn package
// and does not connect to a database. Some Sourcegraph services do not have database access, and
// this defines the entrypoint function for those enterprise services.
package nodbcmd

import (
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
