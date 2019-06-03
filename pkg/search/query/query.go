// Copyright 2016 Google Inc. All rights reserved.
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

// Package query provides a parser and AST for search expressions.
package query

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"regexp/syntax"
	"sort"
	"strings"
)

var _ = log.Println

// Q is a representation for a possibly hierarchical search query.
type Q interface {
	String() string
}

// RegexpQuery is a query looking for regular expressions matches.
type Regexp struct {
	Regexp        *syntax.Regexp
	FileName      bool
	Content       bool
	CaseSensitive bool
}

// Symbol finds a string that is a symbol.
type Symbol struct {
	Atom *Substring
}

func (q *Symbol) String() string {
	return fmt.Sprintf("sym:%s", q.Atom)
}

func (q *Regexp) String() string {
	pref := ""
	if q.FileName {
		pref = "file_"
	}
	if q.CaseSensitive {
		pref = "case_" + pref
	}
	return fmt.Sprintf("%sregex:%q", pref, q.Regexp.String())
}

// gobRegexp wraps Regexp to make it gob-encodable/decodable. Regexp contains syntax.Regexp, which
// contains slices/arrays with possibly nil elements, which gob doesn't support
// (https://github.com/golang/go/issues/1501).
type gobRegexp struct {
	Regexp       // Regexp.Regexp (*syntax.Regexp) is set to nil and its string is set in RegexpString
	RegexpString string
}

// GobEncode implements gob.Encoder.
func (q Regexp) GobEncode() ([]byte, error) {
	gobq := gobRegexp{Regexp: q, RegexpString: q.Regexp.String()}
	gobq.Regexp.Regexp = nil // can't be gob-encoded/decoded
	return json.Marshal(gobq)
}

// GobDecode implements gob.Decoder.
func (q *Regexp) GobDecode(data []byte) error {
	var gobq gobRegexp
	err := json.Unmarshal(data, &gobq)
	if err != nil {
		return err
	}
	gobq.Regexp.Regexp, err = syntax.Parse(gobq.RegexpString, regexpFlags)
	if err != nil {
		return err
	}
	*q = gobq.Regexp
	return nil
}

type caseQ struct {
	Flavor string
}

func (c *caseQ) String() string {
	return "case:" + c.Flavor
}

type Language struct {
	Language string
}

func (l *Language) String() string {
	return "lang:" + l.Language
}

type Const struct {
	Value bool
}

func (q *Const) String() string {
	if q.Value {
		return "TRUE"
	}
	return "FALSE"
}

type Repo struct {
	Pattern string
}

func (q *Repo) String() string {
	return fmt.Sprintf("repo:%s", q.Pattern)
}

// RepoSet is a list of repos to match. It is a Sourcegraph addition and only
// used in the Rest interface for efficient checking of large repo lists.
type RepoSet struct {
	Set map[string]struct{}
}

func (q *RepoSet) String() string {
	var detail string
	if len(q.Set) > 5 {
		// Large sets being output are not useful
		detail = fmt.Sprintf("size=%d", len(q.Set))
	} else {
		repos := make([]string, len(q.Set))
		i := 0
		for repo := range q.Set {
			repos[i] = repo
			i++
		}
		sort.Strings(repos)
		detail = strings.Join(repos, " ")
	}
	return fmt.Sprintf("(reposet %s)", detail)
}

func NewRepoSet(repo ...string) *RepoSet {
	s := &RepoSet{Set: make(map[string]struct{})}
	for _, r := range repo {
		s.Set[r] = struct{}{}
	}
	return s
}

const (
	TypeFileMatch uint8 = iota
	TypeFileName
	TypeRepo
)

// Type changes the result type returned.
type Type struct {
	Child Q
	Type  uint8
}

func (q *Type) String() string {
	switch q.Type {
	case TypeFileMatch:
		return fmt.Sprintf("(type:filematch %s)", q.Child)
	case TypeFileName:
		return fmt.Sprintf("(type:filename %s)", q.Child)
	case TypeRepo:
		return fmt.Sprintf("(type:repo %s)", q.Child)
	default:
		return fmt.Sprintf("(type:UNKNOWN %s)", q.Child)
	}
}

// Substring is the most basic query: a query for a substring.
type Substring struct {
	Pattern       string
	CaseSensitive bool

	// Match only filename
	FileName bool

	// Match only content
	Content bool
}

func (q *Substring) String() string {
	s := ""

	t := ""
	if q.FileName {
		t = "file_"
	} else if q.Content {
		t = "content_"
	}

	s += fmt.Sprintf("%ssubstr:%q", t, q.Pattern)
	if q.CaseSensitive {
		s = "case_" + s
	}
	return s
}

type setCaser interface {
	setCase(string)
}

func (q *Substring) setCase(k string) {
	switch k {
	case "yes":
		q.CaseSensitive = true
	case "no":
		q.CaseSensitive = false
	case "auto":
		// TODO - unicode
		q.CaseSensitive = (q.Pattern != string(toLower([]byte(q.Pattern))))
	}
}

func (q *Symbol) setCase(k string) {
	q.Atom.setCase(k)
}

func (q *Regexp) setCase(k string) {
	switch k {
	case "yes":
		q.CaseSensitive = true
	case "no":
		q.CaseSensitive = false
	case "auto":
		q.CaseSensitive = (q.Regexp.String() != LowerRegexp(q.Regexp).String())
	}
}

// Or is matched when any of its children is matched.
type Or struct {
	Children []Q
}

func (q *Or) String() string {
	var sub []string
	for _, ch := range q.Children {
		sub = append(sub, ch.String())
	}
	return fmt.Sprintf("(or %s)", strings.Join(sub, " "))
}

// Not inverts the meaning of its child.
type Not struct {
	Child Q
}

func (q *Not) String() string {
	return fmt.Sprintf("(not %s)", q.Child)
}

// And is matched when all its children are.
type And struct {
	Children []Q
}

func (q *And) String() string {
	var sub []string
	for _, ch := range q.Children {
		sub = append(sub, ch.String())
	}
	return fmt.Sprintf("(and %s)", strings.Join(sub, " "))
}

// NewAnd is syntactic sugar for constructing And queries.
func NewAnd(qs ...Q) Q {
	return &And{Children: qs}
}

// NewOr is syntactic sugar for constructing Or queries.
func NewOr(qs ...Q) Q {
	return &Or{Children: qs}
}

// Ref limits search to a specific git ref. In zoekt this is called Branch.
type Ref struct {
	Pattern string
}

func (q *Ref) String() string {
	return fmt.Sprintf("ref:%q", q.Pattern)
}

func queryChildren(q Q) []Q {
	switch s := q.(type) {
	case *And:
		return s.Children
	case *Or:
		return s.Children
	}
	return nil
}

func flattenAndOr(children []Q, typ Q) ([]Q, bool) {
	var flat []Q
	changed := false
	for _, ch := range children {
		ch, subChanged := flatten(ch)
		changed = changed || subChanged
		if reflect.TypeOf(ch) == reflect.TypeOf(typ) {
			changed = true
			subChildren := queryChildren(ch)
			if subChildren != nil {
				flat = append(flat, subChildren...)
			}
		} else {
			flat = append(flat, ch)
		}
	}

	return flat, changed
}

// (and (and x y) z) => (and x y z) , the same for "or"
func flatten(q Q) (Q, bool) {
	switch s := q.(type) {
	case *And:
		if len(s.Children) == 1 {
			return s.Children[0], true
		}
		flatChildren, changed := flattenAndOr(s.Children, s)
		return &And{flatChildren}, changed
	case *Or:
		if len(s.Children) == 1 {
			return s.Children[0], true
		}
		flatChildren, changed := flattenAndOr(s.Children, s)
		return &Or{flatChildren}, changed
	case *Not:
		child, changed := flatten(s.Child)
		return &Not{child}, changed
	case *Type:
		child, changed := flatten(s.Child)
		return &Type{Child: child, Type: s.Type}, changed
	default:
		return q, false
	}
}

func mapQueryList(qs []Q, pre, post func(Q) Q) []Q {
	var neg []Q
	for _, sub := range qs {
		neg = append(neg, Map(sub, pre, post))
	}
	return neg
}

func invertConst(q Q) Q {
	c, ok := q.(*Const)
	if ok {
		return &Const{!c.Value}
	}
	return q
}

func evalAndOrConstants(q Q, children []Q) Q {
	_, isAnd := q.(*And)

	children = mapQueryList(children, nil, evalConstants)

	newCH := children[:0]
	for _, ch := range children {
		c, ok := ch.(*Const)
		if ok {
			if c.Value == isAnd {
				continue
			} else {
				return ch
			}
		}
		newCH = append(newCH, ch)
	}
	if len(newCH) == 0 {
		return &Const{isAnd}
	}
	if isAnd {
		return &And{newCH}
	}
	return &Or{newCH}
}

func evalConstants(q Q) Q {
	switch s := q.(type) {
	case *And:
		return evalAndOrConstants(q, s.Children)
	case *Or:
		return evalAndOrConstants(q, s.Children)
	case *Not:
		ch := evalConstants(s.Child)
		if _, ok := ch.(*Const); ok {
			return invertConst(ch)
		}
		if not, ok := ch.(*Not); ok {
			// --x == x
			return not.Child
		}
		return &Not{ch}
	case *Type:
		ch := evalConstants(s.Child)
		if _, ok := ch.(*Const); ok {
			// If q is the root query, then evaluating this to a const changes
			// the type of result we will return. However, the only case this
			// makes sense is `type:repo TRUE` to return all repos or
			// `type:file TRUE` to return all filenames. For other cases we
			// want to do this constant folding though, so we allow the
			// unexpected behaviour mentioned previously.
			return ch
		}
		return &Type{Child: ch, Type: s.Type}
	case *Substring:
		if len(s.Pattern) == 0 {
			return &Const{true}
		}
	case *Regexp:
		if s.Regexp.Op == syntax.OpEmptyMatch {
			return &Const{true}
		}
	case *Ref:
		if s.Pattern == "" {
			return &Const{true}
		}
	case *RepoSet:
		if len(s.Set) == 0 {
			return &Const{true}
		}
	}
	return q
}

func Simplify(q Q) Q {
	q = evalConstants(q)
	for {
		var changed bool
		q, changed = flatten(q)
		if !changed {
			break
		}
	}

	return q
}

// Map runs pre and post over the q. pre if non-nil is called in prefix
// order. post if non-nil is called in postfix order.
func Map(q Q, pre, post func(q Q) Q) Q {
	if pre != nil {
		q = pre(q)
	}
	switch s := q.(type) {
	case *And:
		q = &And{Children: mapQueryList(s.Children, pre, post)}
	case *Or:
		q = &Or{Children: mapQueryList(s.Children, pre, post)}
	case *Not:
		q = &Not{Child: Map(s.Child, pre, post)}
	case *Type:
		q = &Type{Type: s.Type, Child: Map(s.Child, pre, post)}
	}
	if post != nil {
		q = post(q)
	}
	return q
}

// Expand expands Substr queries into (OR file_substr content_substr)
// queries, and the same for Regexp queries..
func ExpandFileContent(q Q) Q {
	switch s := q.(type) {
	case *Substring:
		if !s.FileName && !s.Content {
			f := *s
			f.FileName = true
			c := *s
			c.Content = true
			return NewOr(&f, &c)
		}
	case *Regexp:
		if !s.FileName && !s.Content {
			f := *s
			f.FileName = true
			c := *s
			c.Content = true
			return NewOr(&f, &c)
		}
	}
	return q
}

// ExpandRepo expands all Repo Q in q with listFn. An error is returned if
// listFn returns an error.
//
// listFn is a function which takes a list of repo patterns and returns all
// repository names that satisfies the patterns. A name satisfies the pattern
// if all include match, and no exclude match.
//
// For example if we have the repos github.com/foo/{bar,baz} and
// github.com/hello/world, an example listFn could do the following:
//
//    listFn([]string{"github.com/foo/.*"}, []string{".*baz.*"})
//     -> []string{"github.com/foo/bar"}
func ExpandRepo(q Q, listFn func(include, exclude []string) (map[string]struct{}, error)) (Q, error) {
	// We want nested ors/ands to be flattened
	q = Simplify(q)

	var retErr error
	list := func(inc, exc []string) Q {
		if retErr != nil {
			return &Const{Value: false}
		}
		q, err := listFn(inc, exc)
		if err != nil {
			retErr = err
			return &Const{Value: false}
		}
		if len(q) == 0 {
			return &Const{Value: false}
		}
		return &RepoSet{Set: q}
	}
	q = Map(q, nil, func(q Q) Q {
		and, ok := q.(*And)
		if !ok {
			return q
		}
		// Children have already been mapped, so safe to modify slice (it
		// should be a copy)
		var inc, exc []string
		children := and.Children[:0]
		for _, cs := range and.Children {
			switch c := cs.(type) {
			case *Repo:
				inc = append(inc, c.Pattern)
			case *Not:
				if r, ok := c.Child.(*Repo); ok {
					exc = append(exc, r.Pattern)
				} else {
					children = append(children, c)
				}
			default:
				children = append(children, c)
			}
		}
		if len(inc) > 0 || len(exc) > 0 {
			children = append(children, list(inc, exc))
		}
		return NewAnd(children...)
	})
	// We may still have Repo queries which are not a child of an And. So we
	// need to naively translate those. First Not then Repo (since Repo can be
	// a child of Not).
	q = Map(q, nil, func(q Q) Q {
		if not, ok := q.(*Not); ok {
			if r, ok := not.Child.(*Repo); ok {
				return list(nil, []string{r.Pattern})
			}
		}
		return q
	})
	q = Map(q, nil, func(q Q) Q {
		if r, ok := q.(*Repo); ok {
			return list([]string{r.Pattern}, nil)
		}
		return q
	})
	return Simplify(q), retErr
}

// IsAtom returns true if q is an atom. An atom is a Q without children Q. For
// example And is not an atom, but Repo is.
func IsAtom(q Q) bool {
	switch q.(type) {
	case *And:
	case *Or:
	case *Not:
	case *Type:
	default:
		return true
	}
	return false
}

// VisitAtoms runs `v` on all atom queries within `q`.
func VisitAtoms(q Q, v func(q Q)) {
	switch s := q.(type) {
	case *And:
		for _, c := range s.Children {
			VisitAtoms(c, v)
		}
	case *Or:
		for _, c := range s.Children {
			VisitAtoms(c, v)
		}
	case *Not:
		VisitAtoms(s.Child, v)
	case *Type:
		VisitAtoms(s.Child, v)
	default:
		v(q)
	}
}

// EvalConstant evaluates q to a constant by visiting each atom with eval. If
// the query simplifies to a constant ok is true.
//
// This is equivalent to combining Map with Simplify, and then checking if the
// resulting query is a Const.
func EvalConstant(q Q, eval func(q Q) (value, ok bool)) (value, ok bool) {
	switch s := q.(type) {
	case *And:
		// We can only return true if every child is sure it is true. If we
		// get one sure false, we can be sure we are false. Otherwise we can't
		// be sure.
		allTrue := true
		for _, c := range s.Children {
			v, ok := EvalConstant(c, eval)
			if ok && !v {
				return false, true
			}
			if !ok {
				allTrue = false
			}
		}
		return allTrue, allTrue
	case *Or:
		// We can return true if one child is sure it is true. We can return
		// false if every child is sure it is false. Otherwise we can't be
		// sure.
		allFalse := true
		for _, c := range s.Children {
			v, ok := EvalConstant(c, eval)
			if ok && v {
				return true, true
			}
			if !ok {
				allFalse = false
			}
		}
		return false, allFalse
	case *Not:
		v, ok := EvalConstant(s.Child, eval)
		return !v, ok
	case *Type:
		return EvalConstant(s.Child, eval)
	case *Const:
		return s.Value, true
	default:
		return eval(q)
	}
}
