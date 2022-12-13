// blobstore is the blobstore server.
package main // import "github.com/sourcegraph/sourcegraph/cmd/blobstore"

import (
	"github.com/sourcegraph/sourcegraph/cmd/blobstore/shared"
	"github.com/sourcegraph/sourcegraph/internal/env"
)

func main() {
	env.Lock()
	env.HandleHelpFlag()

	shared.Main()
}
