package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/precise-code-intel-worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
)

func main() {
	enterprisecmd.DeprecatedSingleServiceMainEnterprise(shared.Service)
}
