package auth

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/lazyregexp"
)

func TestNormalizeUsername(t *testing.T) {
	testCases := []struct {
		in     string
		out    string
		hasErr bool
	}{
		{in: "username", out: "username"},
		{in: "john@gmail.com", out: "john"},
		{in: "john.appleseed@gmail.com", out: "john.appleseed"},
		{in: "john+test@gmail.com", out: "john-test"},
		{in: "this@is@not-an-email", out: "this-is-not-an-email"},
		{in: "user.na$e", out: "user.na-e"},
		{in: "2039f0923f0", out: "2039f0923f0"},
		{in: "john(test)@gmail.com", out: "john-test-"},
		{in: "bob!", out: "bob-"},
		{in: "john_doe", out: "john_doe"},
		{in: "john__doe", out: "john__doe"},
		{in: "_john", out: "_john"},
		{in: "__john", out: "__john"},
		{in: "bob_", out: "bob_"},
		{in: "bob__", out: "bob__"},
		{in: "user_@name", out: "user_"},
		{in: "user_@name", out: "user_"},
		{in: "user_@name", out: "user_"},
		{in: "1", out: "1"},
		{in: "a", out: "a"},
		{in: "a-", out: "a-"},
		{in: "--username-", out: "username-"},
		{in: "bob.!bob", out: "bob-bob"},
		{in: "bob@@bob", out: "bob-bob"},
		{in: "username.", out: "username"},
		{in: ".username", out: "username"},
		{in: "user..name", out: "user-name"},
		{in: "user.-name", out: "user-name"},
		{in: ".", hasErr: true},
		{in: "-", hasErr: true},
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

			if !isValidUsername(out) {
				t.Errorf("Normalization succeeded, but output %q is still not a valid username", out)
			}
		}
	}
}

func isValidUsername(name string) bool {
	return validUsername.MatchString(name) && len(name) <= 255 // (255 is the max length of a username in the database
}

// Equivalent to `^\w(?:\w|[-.](?=\w))*-?$` which we have in the DB constraint, but without a lookahead
var validUsername = lazyregexp.New(`^\w(?:(?:[\w.-]\w|\w)*-?|)$`)
