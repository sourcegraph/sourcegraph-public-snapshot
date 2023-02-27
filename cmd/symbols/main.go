// Command symbols is a service that serves code symbols (functions, variables, etc.) from a repository at a
// specific commit.
package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
	"github.com/sourcegraph/sourcegraph/cmd/symbols/shared"
)

func main() {
	osscmd.DeprecatedSingleServiceMainOSS(shared.Service)
}
