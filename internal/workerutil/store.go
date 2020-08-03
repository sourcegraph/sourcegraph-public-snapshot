package workerutil

import "github.com/sourcegraph/sourcegraph/internal/workerutil/store"

type Store = store.Store
type Record = store.Record
type StoreOptions = store.StoreOptions

var NewStore = store.NewStore
