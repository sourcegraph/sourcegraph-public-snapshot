package api

import (
	bundles "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/commits"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/gitserver"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func testAPI(store store.Store, bundleManagerClient bundles.BundleManagerClient, gitserverClient gitserver.Client, commitUpdater commits.Updater) CodeIntelAPI {
	// Wrap in observed, as that's how it's used in production
	return NewObserved(New(store, bundleManagerClient, gitserverClient, commitUpdater), &observation.TestContext)
}
