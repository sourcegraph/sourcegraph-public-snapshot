package operations

import (
	"fmt"

	"github.com/grafana/regexp"

	bk "github.com/sourcegraph/sourcegraph/dev/ci/internal/buildkite"
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
	name  string
	items []setItem
}

// setItem represents either an operation or a set (but not both).
type setItem struct {
	op  Operation
	set *Set
}

func toSetItems(ops []Operation) (items []setItem) {
	for _, op := range ops {
		items = append(items, setItem{op: op})
	}
	return items
}

// NewSet instantiates a new set of Operations.
func NewSet(ops ...Operation) *Set {
	return &Set{items: toSetItems(ops)}
}

// NewNamedSet instantiates a set of Operations to be grouped under the given name.
//
// WARNING: two named sets cannot be merged!
func NewNamedSet(name string, ops ...Operation) *Set {
	set := NewSet(ops...)
	set.name = name
	return set
}

// Append adds the given operations to the pipeline. Operations should ONLY be ADDITIVE.
// Do not remove steps after they are added.
func (o *Set) Append(ops ...Operation) {
	o.items = append(o.items, toSetItems(ops)...)
}

// Merge adds the given set of operations to the end of this one.
//
// WARNING: two named sets cannot be merged!
func (o *Set) Merge(set *Set) {
	// In case we get an empty set
	if set.isEmpty() {
		return
	}
	// If set is named, validate
	if set.isNamed() {
		if o.isNamed() {
			panic(fmt.Sprintf("cannot merge two named sets %q and %q", set.name, o.name))
		}
		o.items = append(o.items, setItem{set: set})
	} else {
		o.items = append(o.items, set.items...)
	}
}

// Apply runs all operations over the given Buildkite pipeline.
func (o *Set) Apply(pipeline *bk.Pipeline) {
	for i, item := range o.items {
		if item.op != nil {
			// This is a single operation - apply it on the pipeline.
			item.op(pipeline)
		} else if item.set != nil {
			// This is a named set of operations - generate a Pipeline, apply the set over
			// it, and then add it as a step within the parent Pipeline.
			//
			// We cannot do this if the parent pipeline is also named, but that check
			// already happens on Merge, so we assume this is safe.
			group := &bk.Pipeline{
				Steps: nil,
				Group: bk.Group{
					Key:   item.set.Key(),
					Group: item.set.name,
				},
				BeforeEveryStepOpts: pipeline.BeforeEveryStepOpts,
				AfterEveryStepOpts:  pipeline.AfterEveryStepOpts,
			}
			item.set.Apply(group)
			pipeline.Steps = append(pipeline.Steps, group)
		} else {
			panic(fmt.Sprintf("invalid item at index %d", i))
		}
	}
}

// isEmpty indicates if this set has no items associated with it.
func (o *Set) isEmpty() bool {
	return len(o.items) == 0
}

var nonAlphaNumeric = regexp.MustCompile("[^a-zA-Z0-9]+")

func (o *Set) Key() string {
	return nonAlphaNumeric.ReplaceAllString(o.name, "")
}

func (o *Set) isNamed() bool {
	return o.name != ""
}

// PipelineSetupSetName should be used with NewNamedSets for operations to add to the
// pipeline setup group.
const PipelineSetupSetName = "Pipeline setup"
