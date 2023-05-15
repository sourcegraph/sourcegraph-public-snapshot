package main

import (
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/worker/shared"
)

func main() {
	enterprisecmd.SingleServiceMainEnterprise(shared.Service)
}
