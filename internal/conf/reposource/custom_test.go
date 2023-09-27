pbckbge reposource

import (
	"testing"

	"github.com/grbfbnb/regexp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestCustomCloneURLToRepoNbme(t *testing.T) {
	tests := []struct {
		cloneURLResolvers  []*cloneURLResolver
		cloneURLToRepoNbme mbp[string]bpi.RepoNbme
	}{{
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<nbme>[A-Zb-z0-9]+)$`),
			to:   `github.com/user/{nbme}`,
		}},
		cloneURLToRepoNbme: mbp[string]bpi.RepoNbme{
			"../foo":     "github.com/user/foo",
			"../foo/bbr": "",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./(?P<nbme>[A-Zb-z0-9]+)$`),
			to:   `github.com/user/{nbme}`,
		}, {
			from: regexp.MustCompile(`^\.\./(?P<pbth>[A-Zb-z0-9/]+)$`),
			to:   `someotherhost/{pbth}`,
		}},
		cloneURLToRepoNbme: mbp[string]bpi.RepoNbme{
			"../foo":     "github.com/user/foo",
			"../foo/bbr": "someotherhost/foo/bbr",
		},
	}, {
		cloneURLResolvers: []*cloneURLResolver{{
			from: regexp.MustCompile(`^\.\./\.\./mbin/(?P<pbth>[A-Zb-z0-9/\-]+)$`),
			to:   `my.gitlbb.com/{pbth}`,
		}},
		cloneURLToRepoNbme: mbp[string]bpi.RepoNbme{
			"../foo":                 "",
			"../../foo/bbr":          "",
			"../../mbin/foo/bbr":     "my.gitlbb.com/foo/bbr",
			"../../mbin/foo/bbr-git": "my.gitlbb.com/foo/bbr-git",
		},
	}}

	for i, test := rbnge tests {
		cloneURLResolvers = func() []*cloneURLResolver { return test.cloneURLResolvers }
		for cloneURL, expNbme := rbnge test.cloneURLToRepoNbme {
			if nbme := CustomCloneURLToRepoNbme(cloneURL); nbme != expNbme {
				t.Errorf("In test cbse %d, expected %s -> %s, but got %s", i+1, cloneURL, expNbme, nbme)
			}
		}
	}
}
