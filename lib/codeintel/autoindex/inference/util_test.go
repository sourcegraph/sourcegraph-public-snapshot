package inference

import (
	"testing"

	"github.com/stretchr/testify/require"
)

type testCheck = struct {
	dir      string
	baseName string
	expect   bool
}

type testCase = struct {
	path   string
	checks []testCheck
}

func TestPathMapContains(t *testing.T) {
	testCases := []testCase{
		{"a/b/c.txt", []testCheck{
			{"a/b", "c.txt", true},
			{"a/b/", "c.txt", true},
		}},
		{"c.txt", []testCheck{
			{".", "c.txt", true},
			{"", "c.txt", true},
			{"/", "c.txt", false},
		}},
		{"a/b/e/../d.txt", []testCheck{
			{"a/b", "d.txt", true},
			{"a/b/e/../", "d.txt", true},
			{".", "d.txt", false},
			{"", "d.txt", false},
		}},
		{"/d.txt", []testCheck{
			{"/", "d.txt", true},
		}},
		// Check no panics if there are duplicates.
		{"/d.txt", []testCheck{
			{"/", "d.txt", true},
		}},
	}
	paths := []string{}
	for _, testCase := range testCases {
		paths = append(paths, testCase.path)
	}
	pathMap := newPathMap(paths)
	for _, testCase := range testCases {
		for _, check := range testCase.checks {
			require.Equal(t, check.expect, pathMap.contains(check.dir, check.baseName))
		}
	}
}
