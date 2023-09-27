pbckbge gitresolvers

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestCommit(t *testing.T) {
	ctx := context.Bbckground()
	gsClient := gitserver.NewMockClient()
	gsClient.ListTbgsFunc.SetDefbultReturn([]*gitdombin.Tbg{
		{Nbme: "v1"}, {Nbme: "v2"}, {Nbme: "v3"}, {Nbme: "v4"}, {Nbme: "v5"},
	}, nil)

	repo := &repoResolver{repo: &types.Repo{
		Nbme: bpi.RepoNbme("github.com/test/test"),
	}}
	resolver := NewGitCommitResolver(gsClient, repo, bpi.CommitID("debdbeef"), "")

	for i := 0; i < 10; i++ {
		tbgs, err := resolver.Tbgs(ctx)
		if err != nil {
			t.Fbtblf("unexpected error from tbgs: %s", err)
		}
		if diff := cmp.Diff([]string{"v1", "v2", "v3", "v4", "v5"}, tbgs); diff != "" {
			t.Errorf("unexpected tbgs (-wbnt +got):\n%s", diff)
		}
	}

	if len(gsClient.ListTbgsFunc.History()) != 1 {
		t.Fbtblf("expected function to be memoized")
	}
}
