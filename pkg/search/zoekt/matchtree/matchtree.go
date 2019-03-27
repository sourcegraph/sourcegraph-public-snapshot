// Copyright 2018 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package matchtree implements an expression tree which can be evaluated to
// collect matches.
package matchtree

import (
	"fmt"
	"log"

	"github.com/sourcegraph/sourcegraph/pkg/search/zoekt/query"
)

// ContentProvider is an abstraction to treat matches for names and content
// with the same code.
type ContentProvider interface {
	Data(fileName bool) []byte
}

// A DocIterator iterates over documents in order.
type DocIterator interface {
	// provide the next document where we can may find something
	// interesting.
	NextDoc() uint32

	// clears any per-document state of the docIterator, and
	// prepares for evaluating the given doc. The argument is
	// strictly increasing over time.
	Prepare(nextDoc uint32)
}

// An expression tree coupled with matches. The matchtree has two
// functions:
//
// * it implements boolean combinations (and, or, not)
//
// * it implements shortcuts, where we skip documents (for example: if there
// are no matches for a literal substring, we can be sure there are no regexp
// matches)..
//
// The general process for a given (shard, query) is
//
// - construct MatchTree for the query
//
// - find all different leaf matchTrees (substring, regexp, etc.)
//
// in a loop:
//
//   - find next doc to process using nextDoc
//
//   - evaluate atoms (leaf expressions that match text)
//
//   - evaluate the tree using matches(), storing the result in map.
//
//   - if the complete tree returns (matches() == true) for the document,
//     collect all text matches by looking at leaf matchTrees
//
type MatchTree interface {
	DocIterator

	// returns whether this Matches, and if we are sure.
	Matches(cp ContentProvider, cost int, known map[MatchTree]bool) (match bool, sure bool)
}

// All is a MatchTree which matches every document.
type All struct {
	// mutable
	firstDone bool
	docID     uint32
}

// None is a MatchTree that matches nothing.
type None struct {
	Why string
}

// And returns a MatchTree that matches a document only if all the children
// do.
func And(children ...MatchTree) MatchTree {
	return &and{children: children}
}

type and struct {
	children []MatchTree
}

type or struct {
	children []MatchTree
}

type not struct {
	child MatchTree
}

type typeFile struct {
	child MatchTree
}

// NoVisit doesn't get visited when collecting matches.
type NoVisit struct {
	MatchTree
}

// all prepare methods

func (t *All) Prepare(doc uint32) {
	t.docID = doc
	t.firstDone = true
}

func (t *None) Prepare(uint32) {}

func (t *and) Prepare(doc uint32) {
	for _, c := range t.children {
		c.Prepare(doc)
	}
}

func (t *or) Prepare(doc uint32) {
	for _, c := range t.children {
		c.Prepare(doc)
	}
}

func (t *not) Prepare(doc uint32) {
	t.child.Prepare(doc)
}

func (t *typeFile) Prepare(doc uint32) {
	t.child.Prepare(doc)
}

// nextDoc

func (t *All) NextDoc() uint32 {
	if !t.firstDone {
		return 0
	}
	return t.docID + 1
}

func (t *None) NextDoc() uint32 {
	return maxUInt32
}

func (t *and) NextDoc() uint32 {
	var max uint32
	for _, c := range t.children {
		m := c.NextDoc()
		if m > max {
			max = m
		}
	}
	return max
}

const maxUInt32 = 0xffffffff

func (t *or) NextDoc() uint32 {
	min := uint32(maxUInt32)
	for _, c := range t.children {
		m := c.NextDoc()
		if m < min {
			min = m
		}
	}
	return min
}

func (t *not) NextDoc() uint32 {
	return 0
}

func (t *typeFile) NextDoc() uint32 {
	return t.child.NextDoc()
}

// all String methods

func (t *All) String() string {
	return "all"
}

func (t *None) String() string {
	return fmt.Sprintf("not(%q)", t.Why)
}

func (t *NoVisit) String() string {
	return fmt.Sprintf("novisit(%v)", t.MatchTree)
}

func (t *and) String() string {
	return fmt.Sprintf("and%v", t.children)
}

func (t *or) String() string {
	return fmt.Sprintf("or%v", t.children)
}

func (t *not) String() string {
	return fmt.Sprintf("not(%v)", t.child)
}

func (t *typeFile) String() string {
	return fmt.Sprintf("f(%v)", t.child)
}

// VisitMatchTree visits all atoms.
func VisitMatchTree(t MatchTree, f func(MatchTree)) {
	switch s := t.(type) {
	case *and:
		for _, ch := range s.children {
			VisitMatchTree(ch, f)
		}
	case *or:
		for _, ch := range s.children {
			VisitMatchTree(ch, f)
		}
	case *NoVisit:
		VisitMatchTree(s.MatchTree, f)
	case *not:
		VisitMatchTree(s.child, f)
	case *typeFile:
		VisitMatchTree(s.child, f)
	default:
		f(t)
	}
}

// VisitMatches visits atoms which contains matches for collection. Does not
// visit NoVisit.
func VisitMatches(t MatchTree, known map[MatchTree]bool, f func(MatchTree)) {
	switch s := t.(type) {
	case *and:
		for _, ch := range s.children {
			if known[ch] {
				VisitMatches(ch, known, f)
			}
		}
	case *or:
		for _, ch := range s.children {
			if known[ch] {
				VisitMatches(ch, known, f)
			}
		}
	case *not:
	case *NoVisit:
		// don't collect into negative trees.
	case *typeFile:
		// We will just gather the filename if we do not visit this tree.
	default:
		f(s)
	}
}

func EvalMatchTree(cp ContentProvider, cost int, known map[MatchTree]bool, mt MatchTree) (bool, bool) {
	if v, ok := known[mt]; ok {
		return v, true
	}

	v, ok := mt.Matches(cp, cost, known)
	if ok {
		known[mt] = v
	}

	return v, ok
}

// all matches() methods.

func (t *All) Matches(cp ContentProvider, cost int, known map[MatchTree]bool) (bool, bool) {
	return true, true
}

func (t *None) Matches(cp ContentProvider, cost int, known map[MatchTree]bool) (bool, bool) {
	return false, true
}

func (t *and) Matches(cp ContentProvider, cost int, known map[MatchTree]bool) (bool, bool) {
	sure := true

	for _, ch := range t.children {
		v, ok := EvalMatchTree(cp, cost, known, ch)
		if ok && !v {
			return false, true
		}
		if !ok {
			sure = false
		}
	}
	return true, sure
}

func (t *or) Matches(cp ContentProvider, cost int, known map[MatchTree]bool) (bool, bool) {
	matches := false
	sure := true
	for _, ch := range t.children {
		v, ok := EvalMatchTree(cp, cost, known, ch)
		if ok {
			// we could short-circuit, but we want to use
			// the other possibilities as a ranking
			// signal.
			matches = matches || v
		} else {
			sure = false
		}
	}
	return matches, sure
}

func (t *not) Matches(cp ContentProvider, cost int, known map[MatchTree]bool) (bool, bool) {
	v, ok := EvalMatchTree(cp, cost, known, t.child)
	return !v, ok
}

func (t *typeFile) Matches(cp ContentProvider, cost int, known map[MatchTree]bool) (bool, bool) {
	return EvalMatchTree(cp, cost, known, t.child)
}

func NewMatchTree(q query.Q, atom func(q query.Q) (MatchTree, error)) (MatchTree, error) {
	switch s := q.(type) {
	case *query.And:
		var r []MatchTree
		for _, ch := range s.Children {
			ct, err := NewMatchTree(ch, atom)
			if err != nil {
				return nil, err
			}
			r = append(r, ct)
		}
		return &and{r}, nil
	case *query.Or:
		var r []MatchTree
		for _, ch := range s.Children {
			ct, err := NewMatchTree(ch, atom)
			if err != nil {
				return nil, err
			}
			r = append(r, ct)
		}
		return &or{r}, nil
	case *query.Not:
		ct, err := NewMatchTree(s.Child, atom)
		return &not{
			child: ct,
		}, err

	case *query.Type:
		if s.Type != query.TypeFileName {
			break
		}

		ct, err := NewMatchTree(s.Child, atom)
		if err != nil {
			return nil, err
		}

		return &typeFile{
			child: ct,
		}, nil

	case *query.Const:
		if s.Value {
			return &All{}, nil
		} else {
			return &None{"const"}, nil
		}
	}

	ct, err := atom(q)
	if err != nil {
		return nil, err
	}
	if ct == nil {
		log.Panicf("type %T", q)
	}
	return ct, err
}
