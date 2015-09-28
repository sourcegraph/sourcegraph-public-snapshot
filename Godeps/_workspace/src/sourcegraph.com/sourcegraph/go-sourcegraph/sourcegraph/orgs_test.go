package sourcegraph

import "testing"

func TestOrgSpec(t *testing.T) {
	tests := []struct {
		str  string
		spec OrgSpec
	}{
		{"a", OrgSpec{Org: "a"}},
		{"$1", OrgSpec{UID: 1}},
	}

	for _, test := range tests {
		spec, err := ParseOrgSpec(test.str)
		if err != nil {
			t.Errorf("%q: ParseOrgSpec failed: %s", test.str, err)
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
