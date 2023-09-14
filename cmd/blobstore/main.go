// blobstore is the blobstore server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/blobstore"

import (
	"github.com/sourcegraph/sourcegraph/cmd/blobstore/shared"
	"github.com/sourcegraph/sourcegraph/cmd/sourcegraph/osscmd"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
)

func main() {
	sanitycheck.Pass()
	osscmd.SingleServiceMainOSS(shared.Service)
}
