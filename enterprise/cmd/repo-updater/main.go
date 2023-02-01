package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
)

func main() {
	enterprisecmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
