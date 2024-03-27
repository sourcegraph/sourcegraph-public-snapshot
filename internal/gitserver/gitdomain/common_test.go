package gitdomain

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestMessage(t *testing.T) {
	t.Run("Subject", func(t *testing.T) {
		tests := map[Message]string{
			"hello":                 "hello",
			"hello\n":               "hello",
			"hello\n\n":             "hello",
			"hello\nworld":          "hello",
			"hello\n\nworld":        "hello",
			"hello\n\nworld\nfoo":   "hello",
			"hello\n\nworld\nfoo\n": "hello",
		}
		for input, want := range tests {
			got := input.Subject()
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}
	})
	t.Run("Body", func(t *testing.T) {
		tests := map[Message]string{
			"hello":                 "",
			"hello\n":               "",
			"hello\n\n":             "",
			"hello\nworld":          "world",
			"hello\n\nworld":        "world",
			"hello\n\nworld\nfoo":   "world\nfoo",
			"hello\n\nworld\nfoo\n": "world\nfoo",
		}
		for input, want := range tests {
			got := input.Body()
			if got != want {
				t.Errorf("got %q, want %q", got, want)
			}
		}
	})
}

func TestValidateBranchName(t *testing.T) {
	for _, tc := range []struct {
		name   string
		branch string
		valid  bool
	}{
		{name: "Valid branch", branch: "valid-branch", valid: true},
		{name: "Valid branch with slash", branch: "rgs/valid-branch", valid: true},
		{name: "Valid branch with @", branch: "valid@branch", valid: true},
		{name: "Path component with .", branch: "valid-/.branch", valid: false},
		{name: "Double dot", branch: "valid..branch", valid: false},
		{name: "End with .lock", branch: "valid-branch.lock", valid: false},
		{name: "No space", branch: "valid branch", valid: false},
		{name: "No tilde", branch: "valid~branch", valid: false},
		{name: "No carat", branch: "valid^branch", valid: false},
		{name: "No colon", branch: "valid:branch", valid: false},
		{name: "No question mark", branch: "valid?branch", valid: false},
		{name: "No asterisk", branch: "valid*branch", valid: false},
		{name: "No open bracket", branch: "valid[branch", valid: false},
		{name: "No trailing slash", branch: "valid-branch/", valid: false},
		{name: "No beginning slash", branch: "/valid-branch", valid: false},
		{name: "No double slash", branch: "valid//branch", valid: false},
		{name: "No trailing dot", branch: "valid-branch.", valid: false},
		{name: "Cannot contain @{", branch: "valid@{branch", valid: false},
		{name: "Cannot be @", branch: "@", valid: false},
		{name: "Cannot contain backslash", branch: "valid\\branch", valid: false},
		{name: "head not allowed", branch: "head", valid: false},
		{name: "Head not allowed", branch: "Head", valid: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			valid := ValidateBranchName(tc.branch)
			assert.Equal(t, tc.valid, valid)
		})
	}
}

func TestRefGlobs(t *testing.T) {
	tests := map[string]struct {
		globs   []RefGlob
		match   []string
		noMatch []string
		want    []string
	}{
		"empty": {
			globs:   nil,
			noMatch: []string{"a"},
		},
		"globs": {
			globs:   []RefGlob{{Include: "refs/heads/"}},
			match:   []string{"refs/heads/a", "refs/heads/b/c"},
			noMatch: []string{"refs/tags/t"},
		},
		"excludes": {
			globs: []RefGlob{
				{Include: "refs/heads/"}, {Exclude: "refs/heads/x"},
			},
			match:   []string{"refs/heads/a", "refs/heads/b", "refs/heads/x/c"},
			noMatch: []string{"refs/tags/t", "refs/heads/x"},
		},
		"implicit leading refs/": {
			globs: []RefGlob{{Include: "heads/"}},
			match: []string{"refs/heads/a"},
		},
		"implicit trailing /*": {
			globs:   []RefGlob{{Include: "refs/heads/a"}},
			match:   []string{"refs/heads/a", "refs/heads/a/b"},
			noMatch: []string{"refs/heads/b"},
		},
	}
	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m, err := CompileRefGlobs(test.globs)
			if err != nil {
				t.Fatal(err)
			}
			for _, ref := range test.match {
				if !m.Match(ref) {
					t.Errorf("want match %q", ref)
				}
			}
			for _, ref := range test.noMatch {
				if m.Match(ref) {
					t.Errorf("want no match %q", ref)
				}
			}
		})
	}
}

func TestIsAbsoluteRevision(t *testing.T) {
	yes := []string{"8cb03d28ad1c6a875f357c5d862237577b06e57c", "20697a062454c29d84e3f006b22eb029d730cd00"}
	no := []string{"ref: refs/heads/appsinfra/SHEP-20-review", "master", "HEAD", "refs/heads/master", "20697a062454c29d84e3f006b22eb029d730cd0", "20697a062454c29d84e3f006b22eb029d730cd000", "  20697a062454c29d84e3f006b22eb029d730cd00  ", "20697a062454c29d84e3f006b22eb029d730cd0 "}
	for _, s := range yes {
		if !IsAbsoluteRevision(s) {
			t.Errorf("%q should be an absolute revision", s)
		}
	}
	for _, s := range no {
		if IsAbsoluteRevision(s) {
			t.Errorf("%q should not be an absolute revision", s)
		}
	}
}

func TestRoundTripBlameHunk(t *testing.T) {
	diff := ""

	err := quick.Check(func(startLine, endLine, startByte, endByte uint32, commitID api.CommitID, message, filename string, authorName, authorEmail string, authorDate fuzzTime) bool {
		original := &Hunk{
			StartLine: startLine,
			EndLine:   endLine,
			StartByte: startByte,
			EndByte:   endByte,
			CommitID:  commitID,
			Message:   message,
			Filename:  filename,
			Author: Signature{
				Name:  authorName,
				Email: authorEmail,
				Date:  time.Time(authorDate),
			},
		}
		converted := HunkFromBlameProto(original.ToProto())
		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}, nil)
	if err != nil {
		t.Fatalf("unexpected diff (-want +got):\n%s", diff)
	}
}

func TestRoundTripCommit(t *testing.T) {
	diff := ""

	err := quick.Check(func(id api.CommitID, message Message, parents []api.CommitID, authorName, authorEmail, committerName, committerEmail string, authorDate, committerDate fuzzTime) bool {
		original := &Commit{
			ID:      id,
			Message: message,
			Parents: parents,
			Author: Signature{
				Name:  authorName,
				Email: authorEmail,
				Date:  time.Time(authorDate),
			},
			Committer: &Signature{
				Name:  committerName,
				Email: committerEmail,
				Date:  time.Time(committerDate),
			},
		}
		converted := CommitFromProto(original.ToProto())
		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}, nil)
	if err != nil {
		t.Fatalf("unexpected diff (-want +got):\n%s", diff)
	}
}

type fuzzTime time.Time

func (fuzzTime) Generate(rand *rand.Rand, _ int) reflect.Value {
	// The maximum representable year in RFC 3339 is 9999, so we'll use that as our upper bound.
	maxDate := time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

	ts := time.Unix(rand.Int63n(maxDate.Unix()), rand.Int63n(int64(time.Second)))
	return reflect.ValueOf(fuzzTime(ts))
}

var _ quick.Generator = fuzzTime{}
