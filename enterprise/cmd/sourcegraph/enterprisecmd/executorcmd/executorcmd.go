// Package executorcmd similar to enterprisecmd, except that it has customizations specific to the
// executor command. The executor command (1) does not connect to a database, and so dbconn is a
// a forbidden import and (2) is not just a service (it has commands like `executor install all`)
// which means environment variable configuration is not always present, and as such that must not
// be enforced in a standard way like in our other service cmds.
package executorcmd

import (
	"github.com/sourcegraph/sourcegraph/internal/service"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

var config = svcmain.Config{}

// DeprecatedSingleServiceMainEnterprise is called from the `main` function of a command in the
// enterprise (non-OSS) build to start a single service (such as frontend or gitserver).
//
// DEPRECATED: See svcmain.DeprecatedSingleServiceMain documentation for more info.
func DeprecatedSingleServiceMainEnterprise(service service.Service) {
	svcmain.DeprecatedSingleServiceMain(service, config, false, false)
}
