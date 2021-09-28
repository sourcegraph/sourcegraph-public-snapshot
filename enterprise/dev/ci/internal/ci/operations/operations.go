package operations

import (
	bk "github.com/sourcegraph/sourcegraph/enterprise/dev/ci/internal/buildkite"
)

// Operation defines a function that adds something to a Buildkite pipeline, such as one
// or more Steps.
//
// Functions that return an Operation should never accept Config as an argument - they
// should only accept `changedFiles` or specific, evaluated arguments, and should never
// conditionally add Steps and Operations - they should only use arguments to create
// variations of specific Operations (e.g. with different arguments).
type Operation func(*bk.Pipeline)

// Set is a container for a set of Operations that compose a pipeline.
type Set struct {
	ops []Operation
}

// NewSet instantiates a new set of Operations.
func NewSet(ops []Operation) Set {
	return Set{ops: ops}
}

// Append adds the given operations to the pipeline. Operations should ONLY be ADDITIVE.
// Do not remove steps after they are added.
func (o *Set) Append(ops ...Operation) {
	o.ops = append(o.ops, ops...)
}

// Merge adds the given set of operations to the end of this one.
func (o *Set) Merge(ops *Set) {
	o.ops = append(o.ops, ops.ops...)
}

// Apply runs all operations over the given Buildkite pipeline.
func (o *Set) Apply(pipeline *bk.Pipeline) {
	for _, p := range o.ops {
		p(pipeline)
	}
}
