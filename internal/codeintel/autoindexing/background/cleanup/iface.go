package cleanup

import dbworkerstore "github.com/sourcegraph/sourcegraph/internal/workerutil/dbworker/store"

type AutoIndexingService interface {
	WorkerutilStore() dbworkerstore.Store
	DependencySyncStore() dbworkerstore.Store
	DependencyIndexingStore() dbworkerstore.Store
}
