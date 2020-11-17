// Command repo-updater periodically updates repositories configured in site configuration and serves repository
// metadata from multiple external code hosts.
package main

import (
	"github.com/sourcegraph/sourcegraph/cmd/repo-updater/shared"
	"github.com/sourcegraph/sourcegraph/internal/authz"
)

func main() {
	// Set dummy authz provider to unblock channel for checking permissions in GraphQL APIs.
	// See https://github.com/sourcegraph/sourcegraph/issues/3847 for details.
	authz.SetProviders(true, []authz.Provider{})

	shared.Main(nil)
}
