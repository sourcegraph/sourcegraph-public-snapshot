package reposource

import (
	"testing"

	"github.com/grafana/regexp"

	"github.com/sourcegraph/sourcegraph/internal/api"
)

func TestCustomCloneURLToRepoName(t *testing.T) {
	tests := []struct {
		cloneURLResolvers  []*cloneURLResolver
		cloneURLToRepoName map[string]api.RepoName
	}{{
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<name>[A-Za-z0-9]+)$`),
			to:   `github.com/user/{name}`,
		}},
		cloneURLToRepoName: map[string]api.RepoName{
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
		cloneURLToRepoName: map[string]api.RepoName{
			"../foo":     "github.com/user/foo",
			"../foo/bar": "someotherhost/foo/bar",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./\.\./main/(?P<path>[A-Za-z0-9/\-]+)$`),
			to:   `my.gitlab.com/{path}`,
		}},
		cloneURLToRepoName: map[string]api.RepoName{
			"../foo":                 "",
			"../../foo/bar":          "",
			"../../main/foo/bar":     "my.gitlab.com/foo/bar",
			"../../main/foo/bar-git": "my.gitlab.com/foo/bar-git",
		},
	}}

	for i, test := range tests {
		cloneURLResolvers = func() []*cloneURLResolver { return test.cloneURLResolvers }
		for cloneURL, expName := range test.cloneURLToRepoName {
			if name := CustomCloneURLToRepoName(cloneURL); name != expName {
				t.Errorf("In test case %d, expected %s -> %s, but got %s", i+1, cloneURL, expName, name)
			}
		}
	}
}
