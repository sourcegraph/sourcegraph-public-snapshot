package constructs


// Trait marker for classes that can be depended upon.
//
// The presence of this interface indicates that an object has
// an `IDependable` implementation.
//
// This interface can be used to take an (ordering) dependency on a set of
// constructs. An ordering dependency implies that the resources represented by
// those constructs are deployed before the resources depending ON them are
// deployed.
type IDependable interface {
}

// The jsii proxy for IDependable
type jsiiProxy_IDependable struct {
	_ byte // padding
}

