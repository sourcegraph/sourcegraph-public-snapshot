package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/cody-api/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
)

func main() {
	enterprisecmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
