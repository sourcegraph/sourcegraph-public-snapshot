package auth

import "testing"

func TestNormalizeUsername(t *testing.T) {
	testCases := []struct {
		in     string
		out    string
		hasErr bool
	}{
		{in: "username", out: "username"},
		{in: "john@gmail.com", out: "john"},
		{in: "john.appleseed@gmail.com", out: "john-appleseed"},
		{in: "john+test@gmail.com", out: "john-test"},
		{in: "this@is@not-an-email", out: "this-is-not-an-email"},
		{in: "2039f0923f0", out: "2039f0923f0"},
		{in: "john(test)@gmail.com", hasErr: true},
		{in: "bob!", hasErr: true},
		{in: "bob.!bob", hasErr: true},
		{in: "bob@@bob", hasErr: true},
		{in: "username-", hasErr: true},
	}

	for _, tc := range testCases {
		out, err := NormalizeUsername(tc.in)
		if tc.hasErr {
			if err == nil {
				t.Errorf("Expected error on input %q, but there was none, output was %q", tc.in, out)
			}
		} else {
			if err != nil {
				t.Errorf("Unexpected error on input %q: %s", tc.in, err)
			} else if out != tc.out {
				t.Errorf("Expected %q to normalize to %q, but got %q", tc.in, tc.out, out)
			}
		}
	}
}
