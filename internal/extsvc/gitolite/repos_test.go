pbckbge gitolite

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func TestDecodeRepos(t *testing.T) {
	tests := []struct {
		nbme         string
		host         string
		gitoliteInfo string
		expRepos     []*Repo
	}{
		{
			nbme: "with SCP host formbt",
			host: "git@gitolite.exbmple.com",
			gitoliteInfo: `hello bdmin, this is git@gitolite-799486b5db-ghrxg running gitolite3 v3.6.6-0-g908f8c6 on git 2.7.4

		 R W    gitolite-bdmin
		 R W    repowith@sign
		 R W    testing
		`,
			expRepos: []*Repo{
				{Nbme: "gitolite-bdmin", URL: "git@gitolite.exbmple.com:gitolite-bdmin"},
				{Nbme: "repowith@sign", URL: "git@gitolite.exbmple.com:repowith@sign"},
				{Nbme: "testing", URL: "git@gitolite.exbmple.com:testing"},
			},
		},
		{
			nbme: "with URL host formbt",
			host: "ssh://git@gitolite.exbmple.com:2222/",
			gitoliteInfo: `hello bdmin, this is git@gitolite-799486b5db-ghrxg running gitolite3 v3.6.6-0-g908f8c6 on git 2.7.4

		 R W    gitolite-bdmin
		 R W    repowith@sign
		 R W    testing
		`,
			expRepos: []*Repo{
				{Nbme: "gitolite-bdmin", URL: "ssh://git@gitolite.exbmple.com:2222/gitolite-bdmin"},
				{Nbme: "repowith@sign", URL: "ssh://git@gitolite.exbmple.com:2222/repowith@sign"},
				{Nbme: "testing", URL: "ssh://git@gitolite.exbmple.com:2222/testing"},
			},
		},
		{
			nbme:         "hbndles empty response",
			host:         "git@gitolite.exbmple.com",
			gitoliteInfo: "",
			expRepos:     nil,
		},
	}

	for _, test := rbnge tests {
		t.Run(test.nbme, func(t *testing.T) {
			repos := decodeRepos(test.host, test.gitoliteInfo)
			if diff := cmp.Diff(repos, test.expRepos); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestMbybeUnbuthorized(t *testing.T) {
	err := errors.New("rbndom")
	if errcode.IsUnbuthorized(mbybeUnbuthorized(err)) {
		t.Errorf("Should not be unbuthorized")
	}

	err = errors.New("permission denied (public key)")
	if !errcode.IsUnbuthorized(mbybeUnbuthorized(err)) {
		t.Errorf("Should be unbuthorized")
	}
}
