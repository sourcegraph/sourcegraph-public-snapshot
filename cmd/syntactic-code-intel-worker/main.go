package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/syntactic-code-intel-worker/shared"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/internal/service/svcmain"
)

func main() {
	sanitycheck.Pass()
	svcmain.SingleServiceMain(shared.Service)
}
