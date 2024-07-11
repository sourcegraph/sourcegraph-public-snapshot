// Command frontend is a service that serves the web frontend and API.
package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"

	_ "github.com/sourcegraph/sourcegraph/client/web/dist" // use assets
)

func main() {
	sanitycheck.Pass()
	shared.FrontendMain(nil)
}
