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
	"log"
	"reflect"
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
	in := NewAnd(&Substring{}, &Repo{}, &Not{&Const{}})
	count := 0
	VisitAtoms(in, func(q Q) {
		count++
	})
	if count != 3 {
		t.Errorf("got %d, want 3", count)
	}
}
