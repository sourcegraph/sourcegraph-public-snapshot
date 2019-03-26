package types

//lint:file-ignore SA6004 We rather have a collection of regular expressions.

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/internal/pkg/search/query/syntax"
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

func TestSetValue(t *testing.T) {
	t.Run("quotes certain broken queries", func(t *testing.T) {
		tests := []struct {
			input     string
			want      string
			field     string
			valueType ValueType
			tokenType syntax.TokenType
		}{
			{`foo(`, `"foo("`, "", RegexpType, syntax.TokenLiteral},
			{`foo[`, `"foo["`, "", RegexpType, syntax.TokenLiteral},
		}

		for _, test := range tests {
			expr := &syntax.Expr{
				Pos:       0,
				Not:       false,
				Field:     test.field,
				Value:     test.input,
				ValueType: test.tokenType,
			}
			value := &Value{syntax: expr}
			err := setValue(value, test.input, test.valueType, test.field)
			if err != nil {
				t.Fatalf("unexpected error: %v\n", err)
			}

			if value.String != nil {
				if *value.String != test.want {
					t.Errorf("unexpected value result:\ngot: `%v`\nwant: `%v`\n", *value.String, test.want)
				}
			}

			if value.Regexp != nil {
				if got := regexp.MustCompile(test.want); !reflect.DeepEqual(*value.Regexp, got) {
					t.Errorf("unexpected value result:\ngot: `%v`\nwant: `%v`\n", *value.Regexp, got)
				}
			}
		}
	})
}

// TODO(isaac): remove this test before PR is merged. It's here as a proof.
func TestCheckNeedToQuoteCurlyBraces(t *testing.T) {
	tests := []struct {
		regexp  string
		sub     string
		matches []string
	}{
		{
			regexp:  "foo{",
			sub:     "foo",
			matches: []string{},
		},
		{
			regexp:  "foo{",
			sub:     "foo{",
			matches: []string{"foo{"},
		},
		{
			regexp:  "foo{0,1}",
			sub:     "foo{",
			matches: []string{"foo"},
		},
		{
			regexp:  "foo{0,1}",
			sub:     "",
			matches: []string{},
		},
		{
			regexp:  "foo{0,1}",
			sub:     "foo",
			matches: []string{"foo"},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("regepx %s; sub %s", test.regexp, test.sub), func(t *testing.T) {
			p, _ := regexp.Compile(test.regexp)
			matches := p.FindAllString(test.sub, -1)

			if len(matches) != len(test.matches) {
				t.Fatalf("unexpected FindAllString results\ngot: %v\nwant: %v\n", matches, test.matches)
			}

			for i := range matches {
				if matches[i] != test.matches[i] {
					t.Errorf("unexpected FindAllString results\ngot: %v\nwant: %v\n", matches, test.matches)
				}
			}
		})
	}
}
