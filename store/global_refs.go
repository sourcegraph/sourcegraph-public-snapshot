package store

import (
	"golang.org/x/net/context"
	"sourcegraph.com/sourcegraph/sourcegraph/go-sourcegraph/sourcegraph"
	"sourcegraph.com/sourcegraph/srclib/store/pb"
)

// GlobalRefs defines the interface for getting and listing global ref locations.
type GlobalRefs interface {
	// Get returns the names and ref counts of all repos and files within those repos
	// that refer the given def.
	Get(ctx context.Context, op *sourcegraph.DefsListRefLocationsOp) (*sourcegraph.RefLocationsList, error)

	// Update takes the graph output of a source unit and updates the set of refs in
	// the global ref store that originate from this source unit.
	Update(ctx context.Context, op *pb.ImportOp) error
}
