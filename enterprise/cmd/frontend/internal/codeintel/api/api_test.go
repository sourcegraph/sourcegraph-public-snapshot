package api

import (
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifstore"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/store"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

func testAPI(store store.Store, lsifStore lsifstore.Store, gitserverClient gitserverClient) CodeIntelAPI {
	// Wrap in observed, as that's how it's used in production
	return NewObserved(New(store, lsifStore, gitserverClient), &observation.TestContext)
}
