package env

import "testing"

func Test_envVarNameToCamelCase(t *testing.T) {
	cases := [][2]string{
		{"FOO", "Foo"},
		{"FOO_BAR", "FooBar"},
		{"FOO_BAR_BAZ", "FooBarBaz"},
	}
	for _, c := range cases {
		if got := envVarNameToCamelCase(c[0]); got != c[1] {
			t.Errorf("expected %q, got %q on input %q", c[1], got, c[0])
		}
	}
}
