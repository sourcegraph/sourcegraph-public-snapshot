package libs

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestDirWithoutDot(t *testing.T) {
	testCases := []struct {
		actual   string
		expected string
	}{
		{dirWithoutDot("foo.txt"), ""},
		{dirWithoutDot("foo/bar.txt"), "foo"},
		{dirWithoutDot("foo/baz"), "foo"},
	}

	for _, testCase := range testCases {
		if testCase.actual != testCase.expected {
			t.Errorf("unexpected dirname: want=%s got=%s", testCase.expected, testCase.actual)
		}
	}
}

func TestAncestorDirs(t *testing.T) {
	expectedAncestors := []string{
		"foo/bar/baz",
		"foo/bar",
		"foo",
		"",
	}
	if diff := cmp.Diff(expectedAncestors, ancestorDirs("foo/bar/baz/bonk.txt")); diff != "" {
		t.Errorf("unexpected ancestor dirs (-want +got):\n%s", diff)

	}
}
