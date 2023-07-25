package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/precise-code-intel-worker/shared"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/sourcegraph/enterprisecmd"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
)

func main() {
	sanitycheck.Pass()
	enterprisecmd.SingleServiceMainEnterprise(shared.Service)
}
