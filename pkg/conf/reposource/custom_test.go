package reposource

import (
	"regexp"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/api"
)

func Test_customCloneURLToRepoURI(t *testing.T) {
	tests := []struct {
		cloneURLResolvers []*cloneURLResolver
		cloneURLToURI     map[string]string
	}{{
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<name>[A-Za-z0-9]+)$`),
			to:   `github.com/user/{name}`,
		}},
		cloneURLToURI: map[string]string{
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
		cloneURLToURI: map[string]string{
			"../foo":     "github.com/user/foo",
			"../foo/bar": "someotherhost/foo/bar",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./\.\./main/(?P<path>[A-Za-z0-9/\-]+)$`),
			to:   `my.gitlab.com/{path}`,
		}},
		cloneURLToURI: map[string]string{
			"../foo":                 "",
			"../../foo/bar":          "",
			"../../main/foo/bar":     "my.gitlab.com/foo/bar",
			"../../main/foo/bar-git": "my.gitlab.com/foo/bar-git",
		},
	}}

	for i, test := range tests {
		cloneURLResolvers = test.cloneURLResolvers
		for cloneURL, expURI := range test.cloneURLToURI {
			if uri := customCloneURLToRepoURI(cloneURL); uri != api.RepoURI(expURI) {
				t.Errorf("In test case %d, expected %s -> %s, but got %s", i+1, cloneURL, expURI, uri)
			}
		}
	}
}
