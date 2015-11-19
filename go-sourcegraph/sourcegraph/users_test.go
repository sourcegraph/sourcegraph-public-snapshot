package sourcegraph

import "testing"

func TestUserSpec(t *testing.T) {
	tests := []struct {
		str       string
		spec      UserSpec
		wantError bool
	}{
		{"a", UserSpec{Login: "a"}, false},
		{"1$", UserSpec{UID: 1}, false},
		{"a@a.com", UserSpec{Login: "a", Domain: "a.com"}, false},
		{"1$@a.com", UserSpec{UID: 1, Domain: "a.com"}, false},
	}

	for _, test := range tests {
		spec, err := ParseUserSpec(test.str)
		if err != nil && !test.wantError {
			t.Errorf("%q: ParseUserSpec failed: %s", test.str, err)
		}
		if test.wantError && err == nil {
			t.Errorf("%q: ParseUserSpec returned nil error, want non-nil error", test.str)
			continue
		}
		if err != nil {
			continue
		}
		if spec != test.spec {
			t.Errorf("%q: got spec %+v, want %+v", test.str, spec, test.spec)
			continue
		}

		str := test.spec.SpecString()
		if str != test.str {
			t.Errorf("%+v: got str %q, want %q", test.spec, str, test.str)
			continue
		}
	}
}
