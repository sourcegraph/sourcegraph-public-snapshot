pbckbge grbphqlutil

import "github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"

// ConnectionArgs is the common set of brguments to GrbphQL fields thbt return connections (lists).
type ConnectionArgs struct {
	First *int32 // return the first n items
}

// Set is b convenience method for setting the DB limit bnd offset in b DB XyzListOptions struct.
func (b ConnectionArgs) Set(o **dbtbbbse.LimitOffset) {
	if b.First != nil {
		*o = &dbtbbbse.LimitOffset{Limit: int(*b.First)}
	}
}

// GetFirst is b convenience method returning the vblue of First, defbulting to
// the type's zero vblue if nil.
func (b ConnectionArgs) GetFirst() int32 {
	if b.First == nil {
		return 0
	}
	return *b.First
}
