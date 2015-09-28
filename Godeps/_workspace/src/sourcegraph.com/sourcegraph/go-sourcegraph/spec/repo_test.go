package spec

import (
	"regexp"
	"testing"
)

// repoTestPatterns are used in both TestRepoPattern and
// TestRepoRevPattern.
var repoTestPatterns = []struct {
	input     string
	wantMatch bool
}{
	{"foo", true},
	{"foo/bar", true},
	{"foo.com/bar", true},
	{"foo.com/bar.baz", true},
	{"fo_o.com/bar", true},

	{"", false},
	{"./foo", false},
	{".foo", false},
	{"/foo", false},
	{"foo/", false},
	{"/foo/", false},
	{".foo.com/bar", false},
	{"foo.com/.bar", false},
	{"foo.com//bar", false},
}

func TestRepo(t *testing.T) {
	pat, err := regexp.Compile("^" + RepoPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input     string
		wantMatch bool
	}{
		// Only include cases here that should not be shared with
		// TestRepoRevPattern.
		{"foo@v", false},
	}
	tests = append(tests, repoTestPatterns...)
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

func TestRepoRevPattern(t *testing.T) {
	pat, err := regexp.Compile("^" + RepoRevPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	type test struct {
		input                           string
		wantMatch                       bool
		wantRepo, wantRev, wantCommitID string
	}

	tests := []test{
		// Only include cases here that should not be shared with
		// TestRepoRevPattern.

		{"foo@v", true, "foo", "v", ""},
		{"foo/bar@v", true, "foo/bar", "v", ""},
		{"foo@my/branch/name", true, "foo", "my/branch/name", ""},
		{"foo@xx===" + commitID, true, "foo", "xx", commitID},
		{"foo@bar~10", true, "foo", "bar~10", ""},
		{"foo@bar^10", true, "foo", "bar^10", ""},

		// Possibly undesired matches (since a component starts with a
		// "."), but unlikely to be problematic.
		{"foo@.bar", true, "foo", ".bar", ""},
		{"foo@.my/branch/name", true, "foo", ".my/branch/name", ""},

		{input: "foo@===" + commitID, wantMatch: false},
		{input: "foo@xx===aa", wantMatch: false},
		{input: "foo@xx===", wantMatch: false},
		{input: "foo@my/.branch/name", wantMatch: false},
		{input: "foo@my/branch/.name", wantMatch: false},
	}

	for _, rttest := range repoTestPatterns {
		tests = append(tests, test{
			input:     rttest.input,
			wantMatch: rttest.wantMatch,
			wantRepo:  rttest.input,
		})
	}

	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		repo, rev, commitID, err := ParseRepoRev(test.input)
		if gotErr, wantErr := err != nil, !test.wantMatch; gotErr != wantErr {
			t.Errorf("%q: got err == %v, want error? == %v", test.input, err, wantErr)
		}
		if err == nil {
			if repo != test.wantRepo {
				t.Errorf("%q: got repo == %q, want %q", test.input, repo, test.wantRepo)
			}
			if rev != test.wantRev {
				t.Errorf("%q: got rev == %q, want %q", test.input, rev, test.wantRev)
			}
			if commitID != test.wantCommitID {
				t.Errorf("%q: got commitID == %q, want %q", test.input, commitID, test.wantCommitID)
			}

			str := RepoRevString(repo, rev, commitID)
			if str != test.input {
				t.Errorf("%q: got string %q, want %q", test.input, str, test.input)
			}
		}
	}
}

func TestResolvedRevPattern(t *testing.T) {
	pat, err := regexp.Compile("^" + ResolvedRevPattern + "$")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		input                 string
		wantMatch             bool
		wantRev, wantCommitID string
	}{
		{"v", true, "v", ""},
		{"v", true, "v", ""},
		{"my/branch/name", true, "my/branch/name", ""},
		{"xx===" + commitID, true, "xx", commitID},
		{"bar~10", true, "bar~10", ""},
		{"bar^10", true, "bar^10", ""},

		// Possibly undesired matches (since a component starts with a
		// "."), but unlikely to be problematic.
		{".bar", true, ".bar", ""},
		{".my/branch/name", true, ".my/branch/name", ""},

		{input: "===" + commitID, wantMatch: false},
		{input: "xx===aa", wantMatch: false},
		{input: "xx===", wantMatch: false},
		{input: "my/.branch/name", wantMatch: false},
		{input: "my/branch/.name", wantMatch: false},
	}
	for _, test := range tests {
		match := pat.MatchString(test.input)
		if match != test.wantMatch {
			t.Errorf("%q: got match == %v, want %v", test.input, match, test.wantMatch)
		}

		rev, commitID, err := ParseResolvedRev(test.input)
		if gotErr, wantErr := err != nil, !test.wantMatch; gotErr != wantErr {
			t.Errorf("%q: got err == %v, want error? == %v", test.input, err, wantErr)
		}
		if err == nil {
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
