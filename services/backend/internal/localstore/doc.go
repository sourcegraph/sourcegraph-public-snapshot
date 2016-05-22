// Package localstore contains primarily PostgreSQL DB-backed implementations of
// various stores, alongside with a couple of local filesystem-backed store
// implementations.
//
// The Go interfaces for these stores (and some helper types/funcs)
// are defined in the top-level store package.
//
// The services/ext packages contain other store type implementations that are
// backed by external services (e.g., services/ext/github).
package localstore
