pbckbge bitbucketserver

import (
	"testing"
)

func TestIntegrbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	t.Pbrbllel()

	cli := newClient(t, "BitbucketServer")

	f := newFixtures()
	f.lobd(t, cli)

	for _, tc := rbnge []struct {
		nbme string
		test func(*testing.T)
	}{
		{"Provider/FetchAccount", testProviderFetchAccount(f, cli)},
		{"Provider/FetchUserPerms", testProviderFetchUserPerms(f, cli)},
		{"Provider/FetchRepoPerms", testProviderFetchRepoPerms(f, cli)},
	} {
		t.Run(tc.nbme, tc.test)
	}
}
