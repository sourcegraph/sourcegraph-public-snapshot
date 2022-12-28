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
				"/src/filename",
			},
		},
		// {
		// 	pattern: "directory/path/",
		// 	paths: []string{
		// 		"/directory/path/file",
		// 	},
		// },
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
