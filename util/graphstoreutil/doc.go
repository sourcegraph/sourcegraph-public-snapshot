// Package graphstoreutil instantiates and configures the graphstore
// (which is where the imported srclib data and its indexes are
// stored).
//
// TODO: This should not be a util function. It should be a store, or
// something like that. Currently the graphstore (in this package) is
// initialized differently from how other stores are, and that's not
// good.
package graphstoreutil
