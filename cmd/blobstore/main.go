// blobstore is the blobstore server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/blobstore"

import (
	"github.com/sourcegraph/sourcegraph/cmd/blobstore/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph-oss/osscmd"
)

func main() {
	osscmd.DeprecatedSingleServiceMainOSS(shared.Service)
}
