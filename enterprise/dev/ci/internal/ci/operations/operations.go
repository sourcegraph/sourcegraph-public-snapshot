package operations

import (
	"fmt"
	"regexp"

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
	name  string
	items []setItem
}

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
func NewSet(ops []Operation) *Set {
	return &Set{items: toSetItems(ops)}
}

func NewNamedSet(name string, ops []Operation) *Set {
	return &Set{name: name, items: toSetItems(ops)}
}

// Append adds the given operations to the pipeline. Operations should ONLY be ADDITIVE.
// Do not remove steps after they are added.
func (o *Set) Append(ops ...Operation) {
	o.items = append(o.items, toSetItems(ops)...)
}

// Merge adds the given set of operations to the end of this one.
func (o *Set) Merge(set *Set) {
	if set.name != "" {
		o.items = append(o.items, setItem{set: set})
	} else {
		o.items = append(o.items, set.items...)
	}
}

// Apply runs all operations over the given Buildkite pipeline.
func (o *Set) Apply(pipeline *bk.Pipeline) {
	for i, item := range o.items {
		if item.op != nil {
			item.op(pipeline)
		} else if item.set != nil {
			group := &bk.Pipeline{
				Key:   item.set.Key(),
				Group: item.set.name,
			}
			item.set.Apply(group)
			pipeline.Steps = append(pipeline.Steps, group)
		} else {
			panic(fmt.Sprintf("invalid item at index %d", i))
		}
	}
}

var nonAlphaNumeric = regexp.MustCompile("[^a-zA-Z0-9]+")

func (o *Set) Key() string {
	return nonAlphaNumeric.ReplaceAllString(o.name, "")
}
