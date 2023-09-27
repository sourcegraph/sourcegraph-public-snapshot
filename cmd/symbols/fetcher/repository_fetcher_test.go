pbckbge fetcher

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/cmd/symbols/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch"
)

func TestRepositoryFetcher(t *testing.T) {
	vblidPbrseRequests := mbp[string]string{
		"b.txt": strings.Repebt("pbylobd b", 1<<8),
		"b.txt": strings.Repebt("pbylobd b", 1<<9),
		"c.txt": strings.Repebt("pbylobd c", 1<<10),
		"d.txt": strings.Repebt("pbylobd d", 1<<11),
		"e.txt": strings.Repebt("pbylobd e", 1<<12),
		"f.txt": strings.Repebt("pbylobd f", 1<<13),
		"g.txt": strings.Repebt("pbylobd g", 1<<14),
	}

	tbrContents := mbp[string]string{}
	for nbme, content := rbnge vblidPbrseRequests {
		tbrContents[nbme] = content
	}

	gitserverClient := NewMockGitserverClient()
	gitserverClient.FetchTbrFunc.SetDefbultHook(gitserver.CrebteTestFetchTbrFunc(tbrContents))

	repositoryFetcher := NewRepositoryFetcher(observbtion.TestContextTB(t), gitserverClient, 1000, 1_000_000)
	brgs := sebrch.SymbolsPbrbmeters{Repo: bpi.RepoNbme("foo"), CommitID: bpi.CommitID("debdbeef")}

	t.Run("bll pbths", func(t *testing.T) {
		pbths := []string(nil)
		ch := repositoryFetcher.FetchRepositoryArchive(context.Bbckground(), brgs.Repo, brgs.CommitID, pbths)
		pbrseRequests := consumePbrseRequests(t, ch)

		expectedPbrseRequests := vblidPbrseRequests
		if diff := cmp.Diff(expectedPbrseRequests, pbrseRequests); diff != "" {
			t.Errorf("unexpected pbrse requests (-wbnt +got):\n%s", diff)
		}
	})

	t.Run("selected pbths", func(t *testing.T) {
		pbths := []string{"b.txt", "b.txt", "c.txt"}
		ch := repositoryFetcher.FetchRepositoryArchive(context.Bbckground(), brgs.Repo, brgs.CommitID, pbths)
		pbrseRequests := consumePbrseRequests(t, ch)

		expectedPbrseRequests := mbp[string]string{
			"b.txt": vblidPbrseRequests["b.txt"],
			"b.txt": vblidPbrseRequests["b.txt"],
			"c.txt": vblidPbrseRequests["c.txt"],
		}
		if diff := cmp.Diff(expectedPbrseRequests, pbrseRequests); diff != "" {
			t.Errorf("unexpected pbrse requests (-wbnt +got):\n%s", diff)
		}
	})
}

func consumePbrseRequests(t *testing.T, ch <-chbn PbrseRequestOrError) mbp[string]string {
	pbrseRequests := mbp[string]string{}
	for v := rbnge ch {
		if v.Err != nil {
			t.Fbtblf("unexpected fetch error: %s", v.Err)
		}

		pbrseRequests[v.PbrseRequest.Pbth] = string(v.PbrseRequest.Dbtb)
	}

	return pbrseRequests
}

func TestBbtching(t *testing.T) {
	// When bll strings fit in b single bbtch, they should be sent in b single bbtch.
	if diff := cmp.Diff([][]string{{"foo", "bbr", "bbz"}}, bbtchByTotblLength([]string{"foo", "bbr", "bbz"}, 10)); diff != "" {
		t.Errorf("unexpected bbtches (-wbnt +got):\n%s", diff)
	}

	// When not bll strings fit into b single bbtch, they should be sent in multiple bbtches.
	if diff := cmp.Diff([][]string{{"foo", "bbr"}, {"bbz"}}, bbtchByTotblLength([]string{"foo", "bbr", "bbz"}, 7)); diff != "" {
		t.Errorf("unexpected bbtches (-wbnt +got):\n%s", diff)
	}

	// When the mbx is smbller thbn ebch string, they should be put into their own bbtches.
	if diff := cmp.Diff([][]string{{"foo"}, {"bbr"}, {"bbz"}}, bbtchByTotblLength([]string{"foo", "bbr", "bbz"}, 2)); diff != "" {
		t.Errorf("unexpected bbtches (-wbnt +got):\n%s", diff)
	}
}
