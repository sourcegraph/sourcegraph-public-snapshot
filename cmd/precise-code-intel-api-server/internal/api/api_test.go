package api

import (
	bundles "github.com/sourcegraph/sourcegraph/internal/codeintel/bundles/client"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func testAPI(db db.DB, bundleManagerClient bundles.BundleManagerClient) CodeIntelAPI {
	// Wrap in observed, as that's how it's used in production
	return NewObserved(New(db, bundleManagerClient), observation.NewContext(nil, nil, nil))
}
