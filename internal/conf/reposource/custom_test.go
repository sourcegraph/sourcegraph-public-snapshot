package reposource

import (
	"regexp"
	"testing"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestCustomCloneURLToRepoName(t *testing.T) {
	tests := []struct {
		cloneURLResolvers  []*cloneURLResolver
		cloneURLToRepoName map[string]string
	}{{
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<name>[A-Za-z0-9]+)$`),
			to:   `github.com/user/{name}`,
		}},
		cloneURLToRepoName: map[string]string{
			"../foo":     "github.com/user/foo",
			"../foo/bar": "",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<name>[A-Za-z0-9]+)$`),
			to:   `github.com/user/{name}`,
		}, {
			from: regexp.MustCompile(`^\.\./(?P<path>[A-Za-z0-9/]+)$`),
			to:   `someotherhost/{path}`,
		}},
		cloneURLToRepoName: map[string]string{
			"../foo":     "github.com/user/foo",
			"../foo/bar": "someotherhost/foo/bar",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./\.\./main/(?P<path>[A-Za-z0-9/\-]+)$`),
			to:   `my.gitlab.com/{path}`,
		}},
		cloneURLToRepoName: map[string]string{
			"../foo":                 "",
			"../../foo/bar":          "",
			"../../main/foo/bar":     "my.gitlab.com/foo/bar",
			"../../main/foo/bar-git": "my.gitlab.com/foo/bar-git",
		},
	}}

	for i, test := range tests {
		cloneURLResolvers = func() interface{} { return test.cloneURLResolvers }
		for cloneURL, expName := range test.cloneURLToRepoName {
			if name := CustomCloneURLToRepoName(cloneURL); name != api.RepoName(expName) {
				t.Errorf("In test case %d, expected %s -> %s, but got %s", i+1, cloneURL, expName, name)
			}
		}
	}
}
