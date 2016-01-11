package client

import (
	"log"

	"golang.org/x/net/context"
	"src.sourcegraph.com/sourcegraph/go-sourcegraph/sourcegraph"
)

// Ctx is a context that accesses the configured Sourcegraph endpoint
// with the configured credentials. It should be used for all CLI
// operations.
var Ctx context.Context

// Client returns a Sourcegraph API client configured to use the
// specified endpoints and authentication info.
func Client() *sourcegraph.Client {
	c, err := sourcegraph.NewClientFromContext(Ctx)
	if err != nil {
		log.Fatalf("could not create client: %s", err)
	}
	return c
}
