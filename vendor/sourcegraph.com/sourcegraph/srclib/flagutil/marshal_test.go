package flagutil

import (
	"reflect"
	"testing"
)

func TestMarshalArgs(t *testing.T) {
	type fooGroup struct {
		Bar string `long:"bar"`
	}

	tests := []struct {
		group    interface{}
		wantArgs []string
	}{
		{
			group: &struct {
				Foo bool `short:"f"`
				Bar bool `short:"b"`
			}{Foo: true},
			wantArgs: []string{"-f"},
		},
		{
			group: &struct {
				Foo string `short:"f"`
			}{Foo: "bar"},
			wantArgs: []string{"-f", "bar"},
		},
		{
			group: &struct {
				Foo string `long:"foo"`
			}{Foo: "bar"},
			wantArgs: []string{"--foo", "bar"},
		},
		{
			group: &struct {
				Foo string `short:"f"`
			}{Foo: "bar baz"},
			wantArgs: []string{"-f", "bar baz"},
		},
		{
			group: &struct {
				Foo string `long:"foo"`
			}{Foo: "bar baz"},
			wantArgs: []string{"--foo", "bar baz"},
		},
		{
			group: &struct {
				Foo string `short:"f" long:"foo"`
			}{Foo: "bar"},
			wantArgs: []string{"--foo", "bar"},
		},
		{
			group: &struct {
				Foo []string `long:"foo"`
			}{Foo: []string{"bar", "baz"}},
			wantArgs: []string{"--foo", "bar", "--foo", "baz"},
		},
		{
			group: &struct {
				Foo []string `long:"foo"`
			}{Foo: []string{}},
			wantArgs: nil,
		},
		{
			group: &struct {
				Foo fooGroup `group:"foo"`
			}{Foo: fooGroup{Bar: "x"}},
			wantArgs: []string{"--bar", "x"},
		},
		{
			group: &struct {
				Foo fooGroup `group:"foo" namespace:"foo"`
			}{Foo: fooGroup{Bar: "x"}},
			wantArgs: []string{"--foo.bar", "x"},
		},

		// Omit default-valued flags
		{
			group: &struct {
				S0 string `long:"s0"`
				S1 string `long:"s1" default:"x"`
				S2 string `long:"s2" default:"y"`
				I0 int    `long:"i0"`
				I1 int    `long:"i1" default:"7"`
				I2 int    `long:"i2" default:"8"`
				B1 bool   `long:"b1"`
				B2 bool   `long:"b2"`
			}{
				S1: "x",
				S2: "abc",
				I1: 7,
				I2: 123,
				B2: true,
			},
			wantArgs: []string{"--s2", "abc", "--i2", "123", "--b2"},
		},
	}
	for _, test := range tests {
		args, err := MarshalArgs(test.group)
		if err != nil {
			t.Error(err)
			continue
		}

		if !reflect.DeepEqual(args, test.wantArgs) {
			t.Errorf("got args %v, want %v", args, test.wantArgs)
		}
	}
}
