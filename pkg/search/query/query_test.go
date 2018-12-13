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

package query

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"
)

var _ = log.Println

func TestQueryString(t *testing.T) {
	q := &Or{[]Q{
		&And{[]Q{
			&Substring{Pattern: "hoi"},
			&Not{&Substring{Pattern: "hai"}},
		}}}}
	got := q.String()
	want := `(or (and substr:"hoi" (not substr:"hai")))`

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestSimplify(t *testing.T) {
	type testcase struct {
		in   Q
		want Q
	}

	cases := []testcase{
		{
			in: NewOr(
				NewOr(
					NewAnd(&Substring{Pattern: "hoi"},
						&Not{&Substring{Pattern: "hai"}}),
					NewOr(
						&Substring{Pattern: "zip"},
						&Substring{Pattern: "zap"},
					))),
			want: NewOr(
				NewAnd(
					&Substring{Pattern: "hoi"},
					&Not{&Substring{Pattern: "hai"}}),
				&Substring{Pattern: "zip"},
				&Substring{Pattern: "zap"}),
		},
		{in: &And{}, want: &Const{true}},
		{in: &Or{}, want: &Const{false}},
		{in: NewAnd(&Const{true}, &Const{false}), want: &Const{false}},
		{in: NewOr(&Const{false}, &Const{true}), want: &Const{true}},
		{in: &Not{&Const{true}}, want: &Const{false}},
		{
			in: NewAnd(
				&Substring{Pattern: "byte"},
				&Not{NewAnd(&Substring{Pattern: "byte"})}),
			want: NewAnd(
				&Substring{Pattern: "byte"},
				&Not{&Substring{Pattern: "byte"}}),
		},
	}

	for _, c := range cases {
		got := Simplify(c.in)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("got %s, want %s", got, c.want)
		}
	}
}

func TestMap(t *testing.T) {
	in := NewAnd(&Substring{Pattern: "bla"}, &Not{&Repo{"foo"}})
	out := NewAnd(&Substring{Pattern: "bla"}, &Not{&Const{false}})

	f := func(q Q) Q {
		if _, ok := q.(*Repo); ok {
			return &Const{false}
		}
		return q
	}
	got := Map(in, nil, f)
	if !reflect.DeepEqual(got, out) {
		t.Errorf("got %v, want %v", got, out)
	}
}

// createListFunc returns a listFn (see ExpandRepo docs) which returns a list
// of repositories from repos.
func createListFunc(repos []string) func([]string, []string) (map[string]struct{}, error) {
	return func(inc, exc []string) (map[string]struct{}, error) {
		set := map[string]struct{}{}
		for _, r := range repos {
			set[r] = struct{}{}
		}
		for r := range set {
			for _, re := range inc {
				if matched, err := regexp.MatchString(re, r); err != nil {
					return nil, err
				} else if !matched {
					delete(set, r)
				}
			}
			for _, re := range exc {
				if matched, err := regexp.MatchString(re, r); err != nil {
					return nil, err
				} else if matched {
					delete(set, r)
				}
			}
		}
		return set, nil
	}
}

func TestExpandRepo(t *testing.T) {
	list := createListFunc([]string{"foo", "bar", "baz"})

	cases := map[string]string{
		"r:":                         "(reposet bar baz foo)",
		"r:b":                        "(reposet bar baz)",
		"r:b r:a":                    "(reposet bar baz)",
		"r:b -r:baz":                 "(reposet bar)",
		"-r:f":                       "(reposet bar baz)",
		"r:foo":                      "(reposet foo)",
		"r:foo r:baz":                "FALSE",
		"r:foo test":                 "(and substr:\"test\" (reposet foo))",
		"r:foo test -hello":          "(and substr:\"test\" (not substr:\"hello\") (reposet foo))",
		"(r:ba test (r:b r:a -r:z))": "(and substr:\"test\" (reposet bar))",
	}
	for qStr, want := range cases {
		q, err := Parse(qStr)
		if err != nil {
			t.Fatal(err)
		}
		got, err := ExpandRepo(q, list)
		if err != nil {
			t.Fatal(err)
		}
		if got.String() != want {
			t.Errorf("expandRepo(%q) got %s want %s", qStr, got.String(), want)
		}
	}
}

func TestExpandRepo_error(t *testing.T) {
	list := func(inc, exc []string) (map[string]struct{}, error) {
		return nil, errors.New("fail")
	}
	q, err := Parse("(foo repo:bar) or (baz repo:bam)")
	if err != nil {
		t.Fatal(err)
	}
	q, err = ExpandRepo(q, list)
	if err == nil {
		t.Fatalf("expected error, got %s", q.String())
	}
}

func TestMap_traversal(t *testing.T) {
	value := func(q Q) string {
		switch c := q.(type) {
		case *Repo:
			return c.Pattern
		case *And:
			return "and"
		case *Or:
			return "or"
		default:
			return "unexpected"
		}
	}

	q := NewAnd(
		&Repo{"a"},
		NewOr(&Repo{"b"}, &Repo{"c"}, &Repo{"d"}))

	preWant := []string{"and", "a", "or", "b", "c", "c1", "c2", "d"}
	preGot := []string{}
	pre := func(q Q) Q {
		preGot = append(preGot, value(q))
		if value(q) == "c" {
			// Test premap works. Should appear in final expression and
			// pre/post lists.
			return NewAnd(&Repo{"c1"}, &Repo{"c2"})
		}
		return q
	}

	postWant := []string{"a", "b", "c1", "c2", "and", "d", "or", "and"}
	postGot := []string{}
	post := func(q Q) Q {
		postGot = append(postGot, value(q))
		if value(q) == "b" {
			// Test postmap works. They shouldn't appear anywhere but the
			// final expression.
			return NewAnd(&Repo{"b1"}, &Repo{"b2"})
		}
		return q
	}

	want := "(and repo:a (or (and repo:b1 repo:b2) (and repo:c1 repo:c2) repo:d))"
	q = Map(q, pre, post)

	if q.String() != want {
		t.Errorf("Unexpected Map response\ngot  %s\nwant %s", q.String(), want)
	}
	if !reflect.DeepEqual(preGot, preWant) {
		t.Errorf("Unexpected pre-order traversal\ngot  %#v\nwant %#v", preGot, preWant)
	}
	if !reflect.DeepEqual(postGot, postWant) {
		t.Errorf("Unexpected post-order traversal\ngot  %#v\nwant %#v", postGot, postWant)
	}
}

func TestVisitAtoms(t *testing.T) {
	in := NewAnd(&Substring{}, NewOr(&Repo{}, &Not{&Const{}}))
	count := 0
	VisitAtoms(in, func(q Q) {
		count++
	})
	if count != 3 {
		t.Errorf("got %d, want 3", count)
	}
}

type testEval struct {
	value, ok bool
}

func (q *testEval) String() string {
	if q.ok {
		return fmt.Sprintf("%v", q.value)
	}
	return "?"
}

func TestEvalConstant(t *testing.T) {
	yes := &testEval{value: true, ok: true}
	no := &testEval{value: false, ok: true}
	unsure := &testEval{ok: false}

	cases := []struct {
		Q    Q
		V    bool
		Sure bool
	}{{
		Q:    &Const{},
		V:    false,
		Sure: true,
	}, {
		Q:    &Const{Value: true},
		V:    true,
		Sure: true,
	}, {
		Q:    NewAnd(yes),
		V:    true,
		Sure: true,
	}, {
		Q:    NewAnd(yes, yes),
		V:    true,
		Sure: true,
	}, {
		Q:    NewAnd(yes, no),
		V:    false,
		Sure: true,
	}, {
		Q:    NewAnd(yes, unsure),
		V:    false,
		Sure: false,
	}, {
		Q:    NewAnd(no, unsure),
		V:    false,
		Sure: true,
	}, {
		Q:    NewAnd(unsure, no),
		V:    false,
		Sure: true,
	}, {
		Q:    &Not{unsure},
		V:    true,
		Sure: false,
	}, {
		Q:    &Not{yes},
		V:    false,
		Sure: true,
	}, {
		Q:    &Not{no},
		V:    true,
		Sure: true,
	}}
	for _, c := range cases {
		v, ok := EvalConstant(c.Q, func(q Q) (v, ok bool) {
			e := q.(*testEval)
			return e.value, e.ok
		})
		if v != c.V || ok != c.Sure {
			t.Errorf("EvalConstant(%v) got v=%v ok=%v; want v=%v ok=%v", c.Q, v, ok, c.V, c.Sure)
		}
	}
}
