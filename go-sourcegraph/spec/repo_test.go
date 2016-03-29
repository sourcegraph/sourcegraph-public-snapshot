package spec

import (
	"regexp"
	"testing"
)

func TestRepo(t *testing.T) {
	pat, err := regexp.Compile("^" + RepoPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input     string
		wantMatch bool
	}{
		{"foo", true},
		{"foo/bar", true},
		{"foo.com/bar", true},
		{"foo.com/-bar", true},
		{"foo.com/-bar-", true},
		{"foo.com/bar-", true},
		{"foo.com/.bar", true},
		{"foo.com/bar.baz", true},
		{"fo_o.com/bar", true},
		{".foo", true},
		{"./foo", true},

		{"", false},
		{"/foo", false},
		{"foo/", false},
		{"/foo/", false},
		{"foo.com/-", false},
		{"foo.com/-/bar", false},
		{"-/bar", false},
		{"/-/bar", false},
		{"bar@a", false},
		{"bar@a/b", false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		repo, err := ParseRepo(test.input)
		if gotErr, wantErr := err != nil, !test.wantMatch; gotErr != wantErr {
			t.Errorf("%q: got err == %v, want error? == %v", test.input, err, wantErr)
		}
		if err == nil {
			if repo != test.input {
				t.Errorf("%q: got repo == %q, want %q", test.input, repo, test.input)
			}

			str := RepoString(repo)
			if str != test.input {
				t.Errorf("%q: got string %q, want %q", test.input, str, test.input)
			}
		}
	}
}

const commitID = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"

func TestResolvedRevPattern(t *testing.T) {
	pat, err := regexp.Compile("^" + RevPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input                 string
		wantMatch             bool
		wantRev, wantCommitID string
	}{
		{"v", true, "v", ""},
		{"v/v", true, "v/v", ""},
		{"my/branch/name", true, "my/branch/name", ""},
		{"xx===" + commitID, true, "xx", commitID},
		{"bar~10", true, "bar~10", ""},
		{"bar^10", true, "bar^10", ""},

		{input: "===" + commitID, wantMatch: false},
		{input: "xx===aa", wantMatch: false},
		{input: "xx===", wantMatch: false},
		{input: "-", wantMatch: false},
		{input: "v/-", wantMatch: false},
		{input: "v/-/v", wantMatch: false},
		{input: "-/v", wantMatch: false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		if test.wantMatch {
			rev, commitID := ParseResolvedRev(test.input)
			if rev != test.wantRev {
				t.Errorf("%q: got rev == %q, want %q", test.input, rev, test.wantRev)
			}
			if commitID != test.wantCommitID {
				t.Errorf("%q: got commitID == %q, want %q", test.input, commitID, test.wantCommitID)
			}

			str := ResolvedRevString(rev, commitID)
			if str != test.input {
				t.Errorf("%q: got string %q, want %q", test.input, str, test.input)
			}
		}
	}
}
