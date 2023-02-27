// Command frontend is the enterprise frontend program.
package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
)

func main() {
	enterprisecmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
