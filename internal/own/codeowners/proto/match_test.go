package proto_test

import (
	"reflect"
	"testing"

	codeownerspb "github.com/sourcegraph/sourcegraph/internal/own/codeowners/proto"
)

type testCase struct {
	pattern string
	paths   []string
}

func TestMatch(t *testing.T) {
	cases := []testCase{
		{
			pattern: "filename",
			paths: []string{
				"/filename",
				"/prefix/filename",
			},
		},
		{
			pattern: "directory/path/",
			paths: []string{
				"/directory/path/file",
				"/directory/path/deeply/nested/file",
				"/prefix/directory/path/file",
				"/prefix/directory/path/deeply/nested/file",
			},
		},
		{
			pattern: "directory/*",
			paths: []string{
				"/directory/file",
				"/prefix/directory/another_file",
			},
		},
		{
			pattern: "/toplevelfile",
			paths: []string{
				"/toplevelfile",
			},
		},
	}
	for _, c := range cases {
		for _, path := range c.paths {
			pattern := c.pattern
			owner := []*codeownerspb.Owner{
				{Handle: "foo"},
			}
			file := &codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{Pattern: pattern, Owner: owner},
				},
			}
			got := file.Match(path)
			if !reflect.DeepEqual(got, owner) {
				t.Errorf("want %q to match %q", pattern, path)
			}
		}
	}
}

func TestNoMatch(t *testing.T) {
	cases := []testCase{
		{
			pattern: "filename",
			paths: []string{
				"/prefix_filename_suffix",
				"/src/prefix_filename",
			},
		},
		{
			pattern: "directory/path/",
			paths: []string{
				"/directory/path_and_more/file",
				"/prefix/directory/path_and_more/file",
				// These do not match as the pattern matches anything within
				// the sub-directory tree, but not the directory itself.
				"/directory/path",
				"/prefix/directory/path",
			},
		},
		{
			pattern: "directory/*",
			paths: []string{
				//"/directory/nested/file",
				"/directory/deeply/nested/file",
			},
		},
		{
			pattern: "/toplevelfile",
			paths: []string{
				"/toplevelfile/nested",
				"/notreally/toplevelfile",
			},
		},
	}
	for _, c := range cases {
		for _, path := range c.paths {
			pattern := c.pattern
			owner := []*codeownerspb.Owner{
				{Handle: "foo"},
			}
			file := &codeownerspb.File{
				Rule: []*codeownerspb.Rule{
					{Pattern: pattern, Owner: owner},
				},
			}
			got := file.Match(path)
			if got != nil {
				t.Errorf("want %q not to match %q", pattern, path)
			}
		}
	}
}
