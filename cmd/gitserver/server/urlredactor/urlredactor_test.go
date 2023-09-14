package urlredactor

import (
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/vcs"
)

func TestUrlRedactor(t *testing.T) {
	testCases := []struct {
		url      string
		message  string
		redacted string
	}{
		{
			url:      "http://token@github.com/foo/bar/",
			message:  "fatal: repository 'http://token@github.com/foo/bar/' not found",
			redacted: "fatal: repository 'http://<redacted>@github.com/foo/bar/' not found",
		},
		{
			url:      "http://user:password@github.com/foo/bar/",
			message:  "fatal: repository 'http://user:password@github.com/foo/bar/' not found",
			redacted: "fatal: repository 'http://user:<redacted>@github.com/foo/bar/' not found",
		},
		{
			url:      "http://git:password@github.com/foo/bar/",
			message:  "fatal: repository 'http://git:password@github.com/foo/bar/' not found",
			redacted: "fatal: repository 'http://git:<redacted>@github.com/foo/bar/' not found",
		},
		{
			url:      "http://token@github.com///repo//nick/",
			message:  "fatal: repository 'http://token@github.com/foo/bar/' not found",
			redacted: "fatal: repository 'http://<redacted>@github.com/foo/bar/' not found",
		},
	}
	for _, testCase := range testCases {
		t.Run("", func(t *testing.T) {
			remoteURL, err := vcs.ParseURL(testCase.url)
			if err != nil {
				t.Fatal(err)
			}
			if actual := New(remoteURL).Redact(testCase.message); actual != testCase.redacted {
				t.Fatalf("newUrlRedactor(%q).redact(%q) got %q; want %q", testCase.url, testCase.message, actual, testCase.redacted)
			}
		})
	}
}
