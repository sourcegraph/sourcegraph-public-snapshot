package spec

import (
	"regexp"
	"strings"
	"testing"
)

func TestUserPattern(t *testing.T) {
	pat, err := regexp.Compile("^" + UserPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input     string
		wantMatch bool
		wantError string
		wantUID   uint32
		wantLogin string
	}{
		{"alice", true, "", 0, "alice"},
		{"alice-x", true, "", 0, "alice-x"},
		{"alice.x", true, "", 0, "alice.x"},
		{"alice_x", true, "", 0, "alice_x"},
		{"123$", true, "", 123, ""},

		{input: "", wantMatch: false},
		{input: ".", wantMatch: false},
		{input: "~", wantMatch: false},
		{input: "$1", wantMatch: false},
		{input: "~@", wantMatch: false},
		{input: "1$@", wantMatch: false},
		{input: "999999999999999999999$", wantMatch: true, wantError: "value out of range"},
		{input: "alice@foo.com", wantMatch: false},
		{input: "alice@", wantMatch: false},
		{input: "alice@~", wantMatch: false},
		{input: "alice@.", wantMatch: false},
		{input: "alice@.com", wantMatch: false},
		{input: "alice@com.", wantMatch: false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		uid, login, err := ParseUser(test.input)
		if test.wantError != "" {
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Errorf("%q: got err == %v, want error to contain %q", test.input, err, test.wantError)
			}
			continue
		}
		if gotErr, wantErr := err != nil, !test.wantMatch; gotErr != wantErr {
			t.Errorf("%q: got err == %v, want error? == %v", test.input, err, wantErr)
		}
		if err == nil {
			if uid != test.wantUID {
				t.Errorf("%q: got uid == %d, want %d", test.input, uid, test.wantUID)
			}
			if login != test.wantLogin {
				t.Errorf("%q: got login == %q, want %q", test.input, login, test.wantLogin)
			}

			str := UserString(uid, login)
			if str != test.input {
				t.Errorf("%q: got string %q, want %q", test.input, str, test.input)
			}
		}
	}
}
