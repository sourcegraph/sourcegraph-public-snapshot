// gitserver is the gitserver server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/gitserver"

import (
	"github.com/sourcegraph/sourcegraph/cmd/gitserver/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
)

func main() {
	osscmd.DeprecatedSingleServiceMainOSS(shared.Service)
}
