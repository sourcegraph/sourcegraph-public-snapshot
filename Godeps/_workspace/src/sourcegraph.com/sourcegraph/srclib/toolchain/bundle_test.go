package toolchain

import (
	"reflect"
	"testing"
)

func TestVariant(t *testing.T) {
	tests := []struct {
		variant Variant
		want    string
	}{
		{Variant{}, ""},
		{Variant{"x": "y"}, "x-y"},
		{Variant{"a": "b", "x": "y", "c": "d"}, "a-b_c-d_x-y"},
	}
	for _, test := range tests {
		got := test.variant.String()
		if got != test.want {
			t.Errorf("%v: got String %q, want %q", test.variant, got, test.want)
			continue
		}
		v := ParseVariant(got)
		if !reflect.DeepEqual(v, test.variant) {
			t.Errorf("ParseVariant(%q): got %+v, want %+v", got, v, test.variant)
			continue
		}
	}
}
