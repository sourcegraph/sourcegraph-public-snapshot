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
	"regexp/syntax"
	"testing"
)

func mustParseRE(s string) *syntax.Regexp {
	r, err := syntax.Parse(s, regexpFlags)
	if err != nil {
		log.Panicf("parsing %q: %v", s, err)
	}
	return r
}

func TestParseQuery(t *testing.T) {
	type testcase struct {
		in   string
		want Q
	}

	for _, c := range []testcase{
		{`\bword\b`, &Regexp{Regexp: mustParseRE(`\bword\b`)}},
		{"fi\"le:bla\"", &Substring{Pattern: "file:bla"}},
		{"abc or def", NewOr(&Substring{Pattern: "abc"}, &Substring{Pattern: "def"})},
		{"(abc or def)", NewOr(&Substring{Pattern: "abc"}, &Substring{Pattern: "def"})},
		{"(ppp qqq or rrr sss)", NewOr(
			NewAnd(&Substring{Pattern: "ppp"}, &Substring{Pattern: "qqq"}),
			NewAnd(&Substring{Pattern: "rrr"}, &Substring{Pattern: "sss"}))},
		{"((x) ora b(z(d)))", NewAnd(
			&Regexp{Regexp: mustParseRE("(x)")},
			&Substring{Pattern: "ora"},
			&Regexp{Regexp: mustParseRE("b(z(d))")})},
		{"( )", &Const{Value: true}},
		{"(abc)(de)", &Regexp{Regexp: mustParseRE("(abc)(de)")}},
		{"sub-pixel", &Substring{Pattern: "sub-pixel"}},
		{"abc", &Substring{Pattern: "abc"}},
		{"ABC", &Substring{Pattern: "ABC", CaseSensitive: true}},
		{"\"abc bcd\"", &Substring{Pattern: "abc bcd"}},
		{"abc bcd", NewAnd(
			&Substring{Pattern: "abc"},
			&Substring{Pattern: "bcd"})},
		{"f:fs", &Substring{Pattern: "fs", FileName: true}},
		{"fs", &Substring{Pattern: "fs"}},
		{"-abc", &Not{&Substring{Pattern: "abc"}}},
		{"abccase:yes", &Substring{Pattern: "abccase:yes"}},
		{"file:abc", &Substring{Pattern: "abc", FileName: true}},
		{"branch:pqr", &Ref{Pattern: "pqr"}},
		{"((x) )", &Regexp{Regexp: mustParseRE("(x)")}},
		{"file:helpers\\.go byte", NewAnd(
			&Substring{Pattern: "helpers.go", FileName: true},
			&Substring{Pattern: "byte"})},
		{"(abc def)", NewAnd(
			&Substring{Pattern: "abc"},
			&Substring{Pattern: "def"})},
		{"(abc def", nil},
		{"regex:abc[p-q]", &Regexp{Regexp: mustParseRE("abc[p-q]")}},
		{"aBc[p-q]", &Regexp{Regexp: mustParseRE("aBc[p-q]"), CaseSensitive: true}},
		{"aBc[p-q] case:auto", &Regexp{Regexp: mustParseRE("aBc[p-q]"), CaseSensitive: true}},
		{"repo:go", &Repo{"go"}},

		{"file:\"\"", &Const{true}},
		{"abc.*def", &Regexp{Regexp: mustParseRE("abc.*def")}},
		{"abc\\.\\*def", &Substring{Pattern: "abc.*def"}},
		{"(abc)", &Regexp{Regexp: mustParseRE("(abc)")}},

		{"c:abc", &Substring{Pattern: "abc", Content: true}},
		{"content:abc", &Substring{Pattern: "abc", Content: true}},

		{"lang:c++", &Language{"c++"}},
		{"sym:pqr", &Symbol{&Substring{Pattern: "pqr"}}},
		{"sym:Pqr", &Symbol{&Substring{Pattern: "Pqr", CaseSensitive: true}}},

		// case
		{"abc case:yes", &Substring{Pattern: "abc", CaseSensitive: true}},
		{"abc case:auto", &Substring{Pattern: "abc", CaseSensitive: false}},
		{"ABC case:auto", &Substring{Pattern: "ABC", CaseSensitive: true}},
		{"ABC case:\"auto\"", &Substring{Pattern: "ABC", CaseSensitive: true}},
		{"abc -f:def case:yes", NewAnd(
			&Substring{Pattern: "abc", CaseSensitive: true},
			&Not{Child: &Substring{Pattern: "def", FileName: true, CaseSensitive: true}},
		)},

		// type
		{"type:repo abc", &Type{Type: TypeRepo, Child: &Substring{Pattern: "abc"}}},
		{"type:file abc def", &Type{Type: TypeFileName, Child: NewAnd(&Substring{Pattern: "abc"}, &Substring{Pattern: "def"})}},
		{"(type:repo abc) def", NewAnd(&Type{Type: TypeRepo, Child: &Substring{Pattern: "abc"}}, &Substring{Pattern: "def"})},

		// errors.
		{"\"abc", nil},
		{"\"a\\", nil},
		{"case:foo", nil},

		{"sym:", nil},
		{"abc or", nil},
		{"or abc", nil},
		{"def or or abc", nil},

		{"", &Const{Value: true}},
	} {
		got, err := Parse(c.in)
		if (c.want == nil) != (err != nil) {
			t.Errorf("Parse(%q): error %v, want %v", c.in, err, c.want)
		} else if got != nil {
			if !reflect.DeepEqual(got, c.want) {
				t.Errorf("Parse(%s): got %v want %v", c.in, got, c.want)
			}
		}
	}
}

func TestTokenize(t *testing.T) {
	type testcase struct {
		in   string
		typ  int
		text string
	}

	cases := []testcase{
		{"file:bla", tokFile, "bla"},
		{"file:bla ", tokFile, "bla"},
		{"f:bla ", tokFile, "bla"},
		{"(abc def) ", tokParenOpen, "("},
		{"(abcdef)", tokText, "(abcdef)"},
		{"(abc)(de)", tokText, "(abc)(de)"},
		{"(ab(c)def) ", tokText, "(ab(c)def)"},
		{"(ab\\ def) ", tokText, "(ab\\ def)"},
		{") ", tokParenClose, ")"},
		{"a(bc))", tokText, "a(bc)"},
		{"abc) ", tokText, "abc"},
		{"file:\"bla\"", tokFile, "bla"},
		{"\"file:bla\"", tokText, "file:bla"},
		{"\\", tokError, ""},
		{"o\"r\" bla", tokText, "or"},
		{"or bla", tokOr, "or"},
		{"ar bla", tokText, "ar"},
	}
	for _, c := range cases {
		tok, err := nextToken([]byte(c.in))
		if err != nil {
			tok = &token{Type: tokError}
		}
		if tok.Type != c.typ {
			t.Errorf("%s: got type %d, want %d", c.in, tok.Type, c.typ)
			continue
		}

		if string(tok.Text) != c.text {
			t.Errorf("%s: got text %q, want %q", c.in, tok.Text, c.text)
		}
	}
}
