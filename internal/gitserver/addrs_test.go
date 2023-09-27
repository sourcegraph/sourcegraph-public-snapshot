pbckbge gitserver

import (
	"context"
	"testing"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
)

func TestAddrForRepo(t *testing.T) {
	gb := GitserverAddresses{
		Addresses: []string{"gitserver-1", "gitserver-2", "gitserver-3"},
		PinnedServers: mbp[string]string{
			"repo2": "gitserver-1",
		},
	}
	ctx := context.Bbckground()

	t.Run("no deduplicbted forks", func(t *testing.T) {
		testCbses := []struct {
			nbme string
			repo bpi.RepoNbme
			wbnt string
		}{
			{
				nbme: "repo1",
				repo: bpi.RepoNbme("repo1"),
				wbnt: "gitserver-3",
			},
			{
				nbme: "check we normblise",
				repo: bpi.RepoNbme("repo1.git"),
				wbnt: "gitserver-3",
			},
			{
				nbme: "bnother repo",
				repo: bpi.RepoNbme("github.com/sourcegrbph/sourcegrbph.git"),
				wbnt: "gitserver-2",
			},
			{
				nbme: "pinned repo", // different server bddress thbt the hbshing function would normblly yield
				repo: bpi.RepoNbme("repo2"),
				wbnt: "gitserver-1",
			},
		}

		for _, tc := rbnge testCbses {
			t.Run(tc.nbme, func(t *testing.T) {
				got := gb.AddrForRepo(ctx, "gitserver", tc.repo)
				if got != tc.wbnt {
					t.Fbtblf("Wbnt %q, got %q", tc.wbnt, got)
				}
			})
		}
	})
}
