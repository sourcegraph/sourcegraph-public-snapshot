package search

import (
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/zoekt"
)

func TestGroupFilesByRepo(t *testing.T) {
	testMatch := func(name, repo string) zoekt.FileMatch {
		return zoekt.FileMatch{
			Repository: repo,
			FileName:   name,
		}
	}

	cases := []struct {
		name     string
		input    []zoekt.FileMatch
		expected map[string][]zoekt.FileMatch
	}{
		{"empty", []zoekt.FileMatch{}, map[string][]zoekt.FileMatch{}},
		{"nil", nil, map[string][]zoekt.FileMatch{}},
		{
			"basic",
			[]zoekt.FileMatch{
				testMatch("file1", "repo1"),
				testMatch("file2", "repo1"),
				testMatch("file3", "repo2"),
			},
			map[string][]zoekt.FileMatch{
				"repo1": {
					testMatch("file1", "repo1"),
					testMatch("file2", "repo1"),
				},
				"repo2": {
					testMatch("file3", "repo2"),
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			output := groupFilesByRepo(tc.input)
			if !reflect.DeepEqual(output, tc.expected) {
				t.Fatal(cmp.Diff(output, tc.expected))
			}
		})
	}
}
