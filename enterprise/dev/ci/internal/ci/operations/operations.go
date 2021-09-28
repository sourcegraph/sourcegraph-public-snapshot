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

// Operations is a container for a set of Operations that compose a pipeline.
type Operations struct {
	ops []Operation
}

// NewOperations instantiates a new set of Operations.
func NewOperations(ops []Operation) Operations {
	return Operations{ops: ops}
}

// Append adds the given operations to the pipeline. Operations should ONLY be ADDITIVE.
// Do not remove steps after they are added.
func (o *Operations) Append(ops ...Operation) {
	o.ops = append(o.ops, ops...)
}

// Merge adds the given set of operations to the end of this one.
func (o *Operations) Merge(ops *Operations) {
	o.ops = append(o.ops, ops.ops...)
}

// Apply runs all operations over the given Buildkite pipeline.
func (o *Operations) Apply(pipeline *bk.Pipeline) {
	for _, p := range o.ops {
		p(pipeline)
	}
}
