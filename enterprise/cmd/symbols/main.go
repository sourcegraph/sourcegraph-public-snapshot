package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/symbols/shared"
)

func main() {
	enterprisecmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
