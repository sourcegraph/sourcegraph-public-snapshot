package graph

import (
	"fmt"
	"regexp"
)

var treePathRegexp = regexp.MustCompile(`^(?:[^/]+)(?:/[^/]+)*$`)

func IsValidTreePath(treePath string) bool {
	return treePathRegexp.MatchString(treePath)
}

func (s *Def) Fmt() DefPrintFormatter { return PrintFormatter(s) }

func (s *Def) sortKey() string { return s.DefKey.String() }

// Propagate describes type/value propagation in code. A Propagate entry from A
// (src) to B (dst) indicates that the type/value of A propagates to B. In Tern,
// this is indicated by A having a "fwd" property whose value is an array that
// includes B.
//
//
// ## Motivation & example
//
// For example, consider the following JavaScript code:
//
//   var a = Foo;
//   var b = a;
//
// Foo, a, and b are each their own def. We could resolve all of them to the
// def of their original type (perhaps Foo), but there are occasions when you
// do want to see only the definition of a or b and examples thereof. Therefore,
// we need to represent them as distinct defs.
//
// Even though Foo, a, and b are distinct defs, there are propagation
// relationships between them that are important to represent. The type of Foo
// propagates to both a and b, and the type of a propagates to b. In this case,
// we would have 3 Propagates: Propagate{Src: "Foo", Dst: "a"}, Propagate{Src:
// "Foo", Dst: "b"}, and Propagate{Src: "a", Dst: "b"}. (The propagation
// relationships could be described by just the first and last Propagates, but
// we explicitly include all paths as a denormalization optimization to avoid
// requiring an unbounded number of DB queries to determine which defs a type
// propagates to or from.)
//
//
// ## Directionality
//
// Propagation is unidirectional, in the general case. In the example above, if
// Foo referred to a JavaScript object and if the code were evaluated, any
// *runtime* type changes (e.g., setting a property) on Foo, a, and b would be
// reflected on all of the others. But this doesn't hold for static analysis;
// it's not always true that if a property "a.x" or "b.x" exists, then "Foo.x"
// exists. The simplest example is when Foo is an external definition. Perhaps
// this example file (which uses Foo as a library) modifies Foo to add a new
// property, but other libraries that use Foo would never see that property
// because they wouldn't be executed in the same context as this example file.
// So, in general, we cannot say that Foo receives all types applied to defs
// that Foo propagates to.
//
//
// ## Hypothetical Python example
//
// Consider the following 2 Python files:
//
//   """file1.py"""
//   class Foo(object): end
//
//   """file2.py"""
//   from .file1 import Foo
//   Foo2 = Foo
//
// In this example, there would be one Propagate: Propagate{Src: "file1/Foo",
// Dst: "file2/Foo2}.
type Propagate struct {
	// Src is the def whose type/value is being propagated to the dst def.
	SrcRepo     string
	SrcPath     string
	SrcUnit     string
	SrcUnitType string

	// Dst is the def that is receiving a propagated type/value from the src def.
	DstRepo     string
	DstPath     string
	DstUnit     string
	DstUnitType string
}

// Sorting

type Defs []*Def

func (vs Defs) Len() int           { return len(vs) }
func (vs Defs) Swap(i, j int)      { vs[i], vs[j] = vs[j], vs[i] }
func (vs Defs) Less(i, j int) bool { return vs[i].sortKey() < vs[j].sortKey() }

func (defs Defs) Keys() (keys []DefKey) {
	keys = make([]DefKey, len(defs))
	for i, def := range defs {
		keys[i] = def.DefKey
	}
	return
}

func (defs Defs) KeySet() (keys map[DefKey]struct{}, err error) {
	keys = make(map[DefKey]struct{})
	for _, def := range defs {
		if _, in := keys[def.DefKey]; in {
			return nil, fmt.Errorf("duplicate def key %+v", def.DefKey)
		}
		keys[def.DefKey] = struct{}{}
	}
	return keys, nil
}
