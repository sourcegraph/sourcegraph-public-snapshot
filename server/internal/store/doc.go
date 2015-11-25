// Package store contains implementations of store types using various
// underlying data sources (filesystem, PostgreSQL, etc.).
//
// The Go interfaces for these stores (and some helper types/funcs)
// are defined in the top-level store package.
//
// The /ext packages contain other store type implementations that are
// backed by external services (e.g., /ext/github and /ext/aws).
//
// Packages under the ./shared directory are e.g. utility functions which are
// shared by multiple store implementations.
package store
