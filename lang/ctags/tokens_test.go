package ctags

import (
	"reflect"
	"testing"
)

func Test_tokenize(t *testing.T) {
	type testcase struct {
		in  string
		out []tokenInfo
	}

	tests := []testcase{
		{in: `foo`, out: []tokenInfo{{"foo", tokName}}},
		{in: `foo bar`, out: []tokenInfo{{"foo", tokName}, {" ", tokWhitespace}, {"bar", tokName}}},
		{in: `Foo::Bar`, out: []tokenInfo{{"Foo", tokName}, {"::", tokOther}, {"Bar", tokName}}},
		{in: `@foo`, out: []tokenInfo{{"@foo", tokName}}},
		{in: `$foo`, out: []tokenInfo{{"$foo", tokName}}},
		{in: `x = foo.$bar?`, out: []tokenInfo{
			{"x", tokName}, {" ", tokWhitespace}, {"=", tokOther}, {" ", tokWhitespace}, {"foo", tokName}, {".", tokOther}, {"$bar?", tokName},
		}},
		{in: "var Δ", out: []tokenInfo{{"var", tokName}, {" ", tokWhitespace}, {"Δ", tokName}}},
	}

	for _, test := range tests {
		got, err := tokenize(test.in)
		if err != nil {
			t.Fatalf("unexpected error on input `%s`: %s", test.in, err)
		}
		if !reflect.DeepEqual(got, test.out) {
			t.Errorf("with input `%s`, got %q, but wanted %q", test.in, got, test.out)
		}
	}
}
