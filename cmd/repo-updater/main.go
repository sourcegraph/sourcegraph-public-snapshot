// Command repo-updater periodically updates repositories configured in site configuration and serves repository
// metadata from multiple external code hosts.
package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
)

func main() {
	osscmd.DeprecatedSingleServiceMainOSS(shared.Service)
}
