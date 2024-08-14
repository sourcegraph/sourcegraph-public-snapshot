package idf

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestTokenizeCamelCase(t *testing.T) {
	type testCase struct {
		s       string
		expToks []string
	}
	cases := []testCase{
		{
			s:       "FooBar",
			expToks: []string{"Foo", "Bar"},
		},
		{
			s:       "fooBarBaz",
			expToks: []string{"foo", "Bar", "Baz"},
		},
		{
			s:       "HTMLParser",
			expToks: []string{"HTML", "Parser"},
		},
		{
			s:       "parseHTML",
			expToks: []string{"parse", "HTML"},
		},
		{
			s:       "HTML5Parser",
			expToks: []string{"HTML5", "Parser"},
		},
		{
			s:       "parseHTML5",
			expToks: []string{"parse", "HTML5"},
		},
	}
	for _, c := range cases {
		toks := tokenizeCamelCase(c.s)
		if diff := cmp.Diff(toks, c.expToks); diff != "" {
			t.Errorf(diff)
		}
	}
}

func TestTokenize(t *testing.T) {
	type testCase struct {
		s       string
		expToks []string
	}
	cases := []testCase{
		{
			s:       "camelCase.snake_case + _weird_.",
			expToks: []string{"camel", "Case", "snake", "case", "weird"},
		},
		{
			s:       "two words camelCase--!:@withPunctuation and_snake_case",
			expToks: []string{"two", "words", "camel", "Case", "with", "Punctuation", "and", "snake", "case"},
		},
	}
	for _, c := range cases {
		toks := Tokenize(c.s)
		if diff := cmp.Diff(c.expToks, toks); diff != "" {
			t.Errorf(diff)
		}
	}
}
