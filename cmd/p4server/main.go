// p4server is the p4server server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/p4server"

import (
	"github.com/sourcegraph/sourcegraph/cmd/p4server/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
)

func main() {
	osscmd.DeprecatedSingleServiceMainOSS(shared.Service)
}
