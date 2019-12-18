package types

import (
	"errors"
	"reflect"
	"regexp"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/sourcegraph/sourcegraph/internal/search/query/syntax"
)

type value struct {
	Not   bool
	Value interface{}
}

func TestCheck(t *testing.T) {
	toValue := func(v *Value) value { return value{Not: v.Not(), Value: v.Value()} }
	toTestValueMap := func(fields map[string][]*Value) map[string][]value {
		m := make(map[string][]value, len(fields))
		for f, vs := range fields {
			m[f] = make([]value, len(vs))
			for i, v := range vs {
				m[f][i] = toValue(v)
			}
		}
		return m
	}

	conf := Config{
		FieldTypes: map[string]FieldType{
			"": {
				Literal: RegexpType,
				Quoted:  StringType,
			},
			"r": {
				Literal:   RegexpType,
				Quoted:    RegexpType,
				Negatable: true,
			},
			"b": {
				Literal:  BoolType,
				Quoted:   BoolType,
				Singular: true,
			},
		},
		FieldAliases: map[string]string{
			"f":  "",
			"r2": "r",
		},
	}
	tests := map[string]struct {
		want    map[string][]value
		wantErr *TypeError
	}{
		"":        {want: map[string][]value{}},
		"a":       {want: map[string][]value{"": {{Value: regexp.MustCompile("a")}}}},
		" a ":     {want: map[string][]value{"": {{Value: regexp.MustCompile("a")}}}},
		`"a"`:     {want: map[string][]value{"": {{Value: "a"}}}},
		"/a b/":   {want: map[string][]value{"": {{Value: regexp.MustCompile("a b")}}}},
		"f:a":     {want: map[string][]value{"": {{Value: regexp.MustCompile("a")}}}},
		`f:"a"`:   {want: map[string][]value{"": {{Value: "a"}}}},
		"f:/a/":   {want: map[string][]value{"": {{Value: regexp.MustCompile("/a/")}}}},
		"r:a":     {want: map[string][]value{"r": {{Value: regexp.MustCompile("a")}}}},
		"r2:a":    {want: map[string][]value{"r": {{Value: regexp.MustCompile("a")}}}},
		`r:"a"`:   {want: map[string][]value{"r": {{Value: regexp.MustCompile("a")}}}},
		"r:/a/":   {want: map[string][]value{"r": {{Value: regexp.MustCompile("/a/")}}}},
		"-r:a":    {want: map[string][]value{"r": {{Not: true, Value: regexp.MustCompile("a")}}}},
		"-r2:a":   {want: map[string][]value{"r": {{Not: true, Value: regexp.MustCompile("a")}}}},
		"b:yes":   {want: map[string][]value{"b": {{Value: true}}}},
		"b:no":    {want: map[string][]value{"b": {{Value: false}}}},
		`b:"yes"`: {want: map[string][]value{"b": {{Value: true}}}},
		`a "b" 'cd'`: {want: map[string][]value{"": {
			{Value: regexp.MustCompile("a")},
			{Value: "b"},
			{Value: "cd"},
		}}},
		`f:a f:b`: {want: map[string][]value{"": {
			{Value: regexp.MustCompile("a")},
			{Value: regexp.MustCompile("b")},
		}}},
		"a f:b -r:c b:yes /d/": {
			want: map[string][]value{
				"": {
					{Value: regexp.MustCompile("a")},
					{Value: regexp.MustCompile("b")},
					{Value: regexp.MustCompile("d")},
				},
				"r": {{Not: true, Value: regexp.MustCompile("c")}},
				"b": {{Value: true}},
			},
		},
		`-a`:         {wantErr: &TypeError{Pos: 1, Err: errors.New(`negated terms (-term) are not yet supported`)}},
		`-b:yes`:     {wantErr: &TypeError{Pos: 1, Err: errors.New(`field "b" does not support negation`)}},
		"b:yes b:no": {wantErr: &TypeError{Pos: 6, Err: errors.New(`field "b" may not be used more than once`)}},
		`/a\x/`:      {wantErr: &TypeError{Pos: 1, Err: errors.New("error parsing regexp: invalid escape sequence: `\\x`")}},
		`"\z"`:       {wantErr: &TypeError{Pos: 0, Err: errors.New(`invalid quoted string: "\z"`)}},
		"b:z":        {wantErr: &TypeError{Pos: 0, Err: errors.New(`invalid boolean "z"`)}},
		`b:"z"`:      {wantErr: &TypeError{Pos: 0, Err: errors.New(`invalid boolean "z"`)}},
		"z:a":        {wantErr: &TypeError{Pos: 0, Err: errors.New(`unrecognized field "z"`)}},
	}
	for input, test := range tests {
		t.Run(input, func(t *testing.T) {
			syntaxQuery, err := syntax.Parse(input)
			if err != nil {
				t.Fatal(err)
			}
			query, err := conf.Check(syntaxQuery)
			if err != nil && test.wantErr == nil {
				t.Fatal(err)
			} else if err == nil && test.wantErr != nil {
				t.Fatalf("got err == nil, want %q", test.wantErr)
			} else if test.wantErr != nil && err.Error() != test.wantErr.Error() {
				t.Fatalf("got err == %q, want %q", err, test.wantErr)
			}
			if err != nil {
				return
			}
			if got := toTestValueMap(query.Fields); !reflect.DeepEqual(got, test.want) {
				t.Errorf("fields\ngot  %+v\nwant %+v", got, test.want)
			}
		})
	}
}

func TestUnquoteString(t *testing.T) {
	tests := map[string]string{
		`"ab"`:    "ab",
		"'ab'":    "ab",
		`'a"b'`:   `a"b`,
		`'a\\"b'`: `a\"b`,
	}
	for input, want := range tests {
		t.Run(input, func(t *testing.T) {
			got, err := unquoteString(input)
			if err != nil {
				t.Fatal(err)
			}
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		})
	}
}

func TestRegexpCompile(t *testing.T) {
	t.Run("curly brace", func(t *testing.T) {
		if _, err := regexp.Compile("{"); err != nil {
			t.Error(err)
		}
	})
}

func Test_autoFix(t *testing.T) {
	t.Run("handcrafted cases", func(t *testing.T) {
		tests := []struct {
			pat  string
			want string
		}{
			{"", ""},
			{"a", "a"},
			{"(", `\(`},
			{"[", `\[`},
			// Unclosed curly braces are already valid regular expressions.
			{"{", `{`},
			{"a(", `a\(`},
			{"(a", `(a`},
			{"(a)", "(a)"},
			{`\ba(`, `\ba\(`},
			{`\bfoo(`, `\bfoo\(`},
			{"*", `*`},
			{"*a", `*a`},
			{"a*", "a*"},
			{"$myvar", `$myvar`},
			{"$f(", `$f\(`},
			{"f()", `f\(\)`},
			{"()", `\(\)`},
			{`(\()`, `(\()`},
			{"()f", `\(\)f`},
			{"f(a", `f(a`},
			{"f(a,", `f(a,`},
			{"b)", `b)`},
			{"[)(]", `[)(]`},
			{"[(]", `[(]`},
			{"[()]", `[\(\)]`},

			// From quick check with initial seed 1557931765714982622
			{"()()", `\(\)\(\)`},
			{"}?}${}.?3]}{()()", `}?}${}.?3]}{\(\)\(\)`},
		}
		for _, tt := range tests {
			t.Run(tt.pat, func(t *testing.T) {
				if got := autoFix(tt.pat); got != tt.want {
					t.Errorf("autoFix(`%v`) = `%v`, want `%v`", tt.pat, got, tt.want)
				}
				// Make sure autoFix never causes regexes that compile to no longer compile.
				_, err := regexp.Compile(tt.pat)
				if err == nil {
					if _, err := regexp.Compile(tt.want); err != nil {
						t.Errorf("want %q regexp fails to compile: %s", tt.want, err)
					}
				}
				once := autoFix(tt.pat)
				twice := autoFix(once)
				if twice != once {
					t.Errorf("autoFix is not idempotent. autoFix(autoFix(`%s`)) = `%s` != autoFix(`%s`) = `%s`", tt.pat, twice, tt.pat, once)
				}
			})
		}
	})
	t.Run("quick check", func(t *testing.T) {
		pp := gopter.DefaultTestParameters()
		props := gopter.NewProperties(pp)
		props.Property("idempotent", prop.ForAll(
			func(s string) bool {
				once := autoFix(s)
				twice := autoFix(once)
				return once == twice
			},
			gen.RegexMatch(`^?[abc123()\[\]{}.?\\^$*+]*$?`),
		))
		props.TestingRun(t)
	})
}
