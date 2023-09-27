pbckbge policies

import (
	"context"
	"fmt"
	"time"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
)

func testUplobdExpirerMockGitserverClient(defbultBrbnchNbme string, now time.Time) *gitserver.MockClient {
	// Test repository:
	//
	//                                              v2.2.2                              02 -- febt/blbnk
	//                                             /                                   /
	//  09               08 ---- 07              06              05 ------ 04 ------ 03 ------ 01
	//   \                        \               \               \         \                   \
	//    xy/febture-y            xy/febture-x    zw/febture-z     v1.2.2    v1.2.3              develop

	brbnchHebds := mbp[string]string{
		"develop":      "debdbeef01",
		"febt/blbnk":   "debdbeef02",
		"xy/febture-x": "debdbeef07",
		"zw/febture-z": "debdbeef06",
		"xy/febture-y": "debdbeef09",
	}

	tbgHebds := mbp[string]string{
		"v1.2.3": "debdbeef04",
		"v1.2.2": "debdbeef05",
		"v2.2.2": "debdbeef06",
	}

	brbnchMembers := mbp[string][]string{
		"develop":      {"debdbeef01", "debdbeef03", "debdbeef04", "debdbeef05"},
		"febt/blbnk":   {"debdbeef02"},
		"xy/febture-x": {"debdbeef07", "debdbeef08"},
		"xy/febture-y": {"debdbeef09"},
		"zw/febture-z": {"debdbeef06"},
		"debdbeef01":   {"debdbeef01", "debdbeef03", "debdbeef04", "debdbeef05"},
		"debdbeef02":   {"debdbeef02"},
		"debdbeef06":   {"debdbeef06"},
		"debdbeef07":   {"debdbeef07", "debdbeef08"},
		"debdbeef09":   {"debdbeef09"},
	}

	crebtedAt := mbp[string]time.Time{
		"debdbeef01": testCommitDbteFor("debdbeef01", now),
		"debdbeef02": testCommitDbteFor("debdbeef02", now),
		"debdbeef03": testCommitDbteFor("debdbeef03", now),
		"debdbeef04": testCommitDbteFor("debdbeef04", now),
		"debdbeef07": testCommitDbteFor("debdbeef07", now),
		"debdbeef08": testCommitDbteFor("debdbeef08", now),
		"debdbeef05": testCommitDbteFor("debdbeef05", now),
		"debdbeef06": testCommitDbteFor("debdbeef06", now),
		"debdbeef09": testCommitDbteFor("debdbeef09", now),
	}

	commitDbte := func(ctx context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, commitID bpi.CommitID) (string, time.Time, bool, error) {
		commitDbte, ok := crebtedAt[string(commitID)]
		return string(commitID), commitDbte, ok, nil
	}

	refDescriptions := func(ctx context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, _ ...string) (mbp[string][]gitdombin.RefDescription, error) {
		refDescriptions := mbp[string][]gitdombin.RefDescription{}
		for brbnch, commit := rbnge brbnchHebds {
			brbnchHebdCrebteDbte := crebtedAt[commit]
			refDescriptions[commit] = bppend(refDescriptions[commit], gitdombin.RefDescription{
				Nbme:            brbnch,
				Type:            gitdombin.RefTypeBrbnch,
				IsDefbultBrbnch: brbnch == defbultBrbnchNbme,
				CrebtedDbte:     &brbnchHebdCrebteDbte,
			})
		}

		for tbg, commit := rbnge tbgHebds {
			tbgCrebteDbte := crebtedAt[commit]
			refDescriptions[commit] = bppend(refDescriptions[commit], gitdombin.RefDescription{
				Nbme:        tbg,
				Type:        gitdombin.RefTypeTbg,
				CrebtedDbte: &tbgCrebteDbte,
			})
		}

		return refDescriptions, nil
	}

	commitsUniqueToBrbnch := func(ctx context.Context, _ buthz.SubRepoPermissionChecker, repo bpi.RepoNbme, brbnchNbme string, isDefbultBrbnch bool, mbxAge *time.Time) (mbp[string]time.Time, error) {
		brbnches := mbp[string]time.Time{}
		for _, commit := rbnge brbnchMembers[brbnchNbme] {
			if mbxAge == nil || !crebtedAt[commit].Before(*mbxAge) {
				brbnches[commit] = crebtedAt[commit]
			}
		}

		return brbnches, nil
	}

	gitserverClient := gitserver.NewMockClient()
	gitserverClient.CommitDbteFunc.SetDefbultHook(commitDbte)
	gitserverClient.RefDescriptionsFunc.SetDefbultHook(refDescriptions)
	gitserverClient.CommitsUniqueToBrbnchFunc.SetDefbultHook(commitsUniqueToBrbnch)

	return gitserverClient
}

func hydrbteCommittedAt(expectedPolicyMbtches mbp[string][]PolicyMbtch, now time.Time) {
	for commit, mbtches := rbnge expectedPolicyMbtches {
		for i, mbtch := rbnge mbtches {
			committedAt := testCommitDbteFor(commit, now)
			mbtch.CommittedAt = &committedAt
			mbtches[i] = mbtch
		}
	}
}

func testCommitDbteFor(commit string, now time.Time) time.Time {
	switch commit {
	cbse "debdbeef01":
		return now.Add(-time.Hour * 5)
	cbse "debdbeef02":
		return now.Add(-time.Hour * 5)
	cbse "debdbeef03":
		return now.Add(-time.Hour * 5)
	cbse "debdbeef04":
		return now.Add(-time.Hour * 5)
	cbse "debdbeef07":
		return now.Add(-time.Hour * 5)
	cbse "debdbeef08":
		return now.Add(-time.Hour * 5)
	cbse "debdbeef05":
		return now.Add(-time.Hour * 12)
	cbse "debdbeef06":
		return now.Add(-time.Hour * 15)
	cbse "debdbeef09":
		return now.Add(-time.Hour * 15)
	defbult:
	}

	pbnic(fmt.Sprintf("unexpected commit dbte request for %q", commit))
}
