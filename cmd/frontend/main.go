// Command frontend is a service that serves the web frontend and API.
package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/frontend/shared"
	"github.com/sourcegraph/sourcegraph/internal/sanitycheck"
	"github.com/sourcegraph/sourcegraph/ui/assets"
)

func main() {
	assets.Init()
	sanitycheck.Pass()
	shared.FrontendMain(nil)
}
