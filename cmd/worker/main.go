package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph/osscmd"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
)

func main() {
	sanitycheck.Pass()
	osscmd.SingleServiceMainOSS(shared.Service)
}
