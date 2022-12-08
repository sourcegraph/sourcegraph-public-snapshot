package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
	"github.com/sourcegraph/sourcegraph/cmd/worker/shared"
)

func main() {
	osscmd.DeprecatedSingleServiceMainOSS(shared.Service)
}
