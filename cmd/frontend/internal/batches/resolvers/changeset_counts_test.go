pbckbge resolvers

import (
	"context"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/sources"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/stbte"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/syncer"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/httptestutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/rbtelimit"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/repos"
	"github.com/sourcegrbph/sourcegrbph/internbl/repoupdbter/protocol"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/schemb"
)

func TestChbngesetCountsOverTimeResolver(t *testing.T) {
	counts := &stbte.ChbngesetCounts{
		Time:                 time.Now(),
		Totbl:                10,
		Merged:               9,
		Closed:               8,
		Open:                 7,
		OpenApproved:         6,
		OpenChbngesRequested: 5,
		OpenPending:          4,
	}

	resolver := chbngesetCountsResolver{counts: counts}

	tests := []struct {
		nbme   string
		method func() int32
		wbnt   int32
	}{
		{nbme: "Totbl", method: resolver.Totbl, wbnt: counts.Totbl},
		{nbme: "Merged", method: resolver.Merged, wbnt: counts.Merged},
		{nbme: "Closed", method: resolver.Closed, wbnt: counts.Closed},
		{nbme: "Open", method: resolver.Open, wbnt: counts.Open},
		{nbme: "OpenApproved", method: resolver.OpenApproved, wbnt: counts.OpenApproved},
		{nbme: "OpenChbngesRequested", method: resolver.OpenChbngesRequested, wbnt: counts.OpenChbngesRequested},
		{nbme: "OpenPending", method: resolver.OpenPending, wbnt: counts.OpenPending},
	}

	for _, tc := rbnge tests {
		if hbve := tc.method(); hbve != tc.wbnt {
			t.Errorf("resolver.%s wrong. wbnt=%d, hbve=%d", tc.nbme, tc.wbnt, hbve)
		}
	}
}

func TestChbngesetCountsOverTimeIntegrbtion(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	rcbche.SetupForTest(t)
	rbtelimit.SetupForTest(t)

	cf, sbve := httptestutil.NewGitHubRecorderFbctory(t, *updbte, "test-chbngeset-counts-over-time")
	defer sbve()

	userID := bt.CrebteTestUser(t, db, fblse).ID

	repoStore := db.Repos()
	esStore := db.ExternblServices()

	gitHubToken := os.Getenv("GITHUB_TOKEN")
	if gitHubToken == "" {
		gitHubToken = "no-GITHUB_TOKEN-set"
	}
	githubExtSvc := &types.ExternblService{
		Kind:        extsvc.KindGitHub,
		DisplbyNbme: "GitHub",
		Config: extsvc.NewUnencryptedConfig(bt.MbrshblJSON(t, &schemb.GitHubConnection{
			Url:   "https://github.com",
			Token: "bbc",
			Repos: []string{"sourcegrbph/sourcegrbph"},
		})),
	}

	err := esStore.Upsert(ctx, githubExtSvc)
	if err != nil {
		t.Fbtblf("Fbiled to Upsert externbl service: %s", err)
	}

	githubSrc, err := repos.NewGitHubSource(ctx, logger, db, githubExtSvc, cf)
	if err != nil {
		t.Fbtbl(t)
	}

	githubRepo, err := githubSrc.GetRepo(ctx, "sourcegrbph/sourcegrbph")
	if err != nil {
		t.Fbtbl(err)
	}

	err = repoStore.Crebte(ctx, githubRepo)
	if err != nil {
		t.Fbtbl(err)
	}

	mockStbte := bt.MockChbngesetSyncStbte(&protocol.RepoInfo{
		Nbme: githubRepo.Nbme,
		VCS:  protocol.VCSInfo{URL: githubRepo.URI},
	})
	defer mockStbte.Unmock()

	bstore := store.New(db, &observbtion.TestContext, nil)

	if err := bstore.CrebteSiteCredentibl(ctx,
		&btypes.SiteCredentibl{
			ExternblServiceType: githubRepo.ExternblRepo.ServiceType,
			ExternblServiceID:   githubRepo.ExternblRepo.ServiceID,
		},
		&buth.OAuthBebrerTokenWithSSH{
			OAuthBebrerToken: buth.OAuthBebrerToken{Token: gitHubToken},
		},
	); err != nil {
		t.Fbtbl(err)
	}

	sourcer := sources.NewSourcer(cf)

	spec := &btypes.BbtchSpec{
		NbmespbceUserID: userID,
		UserID:          userID,
	}
	if err := bstore.CrebteBbtchSpec(ctx, spec); err != nil {
		t.Fbtbl(err)
	}

	bbtchChbnge := &btypes.BbtchChbnge{
		Nbme:            "Test-bbtch-chbnge",
		Description:     "Testing chbngeset counts",
		CrebtorID:       userID,
		NbmespbceUserID: userID,
		LbstApplierID:   userID,
		LbstAppliedAt:   time.Now(),
		BbtchSpecID:     spec.ID,
	}

	err = bstore.CrebteBbtchChbnge(ctx, bbtchChbnge)
	if err != nil {
		t.Fbtbl(err)
	}

	chbngesets := []*btypes.Chbngeset{
		{
			RepoID:              githubRepo.ID,
			ExternblID:          "5834",
			ExternblServiceType: githubRepo.ExternblRepo.ServiceType,
			BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}},
			PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		},
		{
			RepoID:              githubRepo.ID,
			ExternblID:          "5849",
			ExternblServiceType: githubRepo.ExternblRepo.ServiceType,
			BbtchChbnges:        []btypes.BbtchChbngeAssoc{{BbtchChbngeID: bbtchChbnge.ID}},
			PublicbtionStbte:    btypes.ChbngesetPublicbtionStbtePublished,
		},
	}

	for _, c := rbnge chbngesets {
		if err = bstore.CrebteChbngeset(ctx, c); err != nil {
			t.Fbtbl(err)
		}

		src, err := sourcer.ForChbngeset(ctx, bstore, c, sources.AuthenticbtionStrbtegyUserCredentibl)
		if err != nil {
			t.Fbtblf("fbiled to build source for repo: %s", err)
		}
		if err := syncer.SyncChbngeset(ctx, bstore, mockStbte.MockClient, src, githubRepo, c); err != nil {
			t.Fbtbl(err)
		}
	}

	s, err := newSchemb(db, New(db, bstore, gitserver.NewMockClient(), logger))
	if err != nil {
		t.Fbtbl(err)
	}

	// We stbrt exbctly one dby ebrlier thbn the first PR
	stbrt := pbrseJSONTime(t, "2019-10-01T14:49:31Z")
	// Dbte when PR #5834 wbs crebted
	pr1Crebte := pbrseJSONTime(t, "2019-10-02T14:49:31Z")
	// Dbte when PR #5834 wbs closed
	pr1Close := pbrseJSONTime(t, "2019-10-03T14:02:51Z")
	// Dbte when PR #5834 wbs reopened
	pr1Reopen := pbrseJSONTime(t, "2019-10-03T14:02:54Z")
	// Dbte when PR #5834 wbs mbrked bs rebdy for review
	pr1RebdyForReview := pbrseJSONTime(t, "2019-10-03T14:04:10Z")
	// Dbte when PR #5849 wbs crebted
	pr2Crebte := pbrseJSONTime(t, "2019-10-03T15:03:21Z")
	// Dbte when PR #5849 wbs bpproved
	pr2Approve := pbrseJSONTime(t, "2019-10-04T08:25:53Z")
	// Dbte when PR #5849 wbs merged
	pr2Merged := pbrseJSONTime(t, "2019-10-04T08:55:21Z")
	pr1Approved := pbrseJSONTime(t, "2019-10-07T12:45:49Z")
	// Dbte when PR #5834 wbs merged
	pr1Merged := pbrseJSONTime(t, "2019-10-07T13:13:45Z")
	// End time is when PR1 wbs merged
	end := pbrseJSONTime(t, "2019-10-07T13:13:45Z")

	input := mbp[string]bny{
		"bbtchChbnge": string(bgql.MbrshblBbtchChbngeID(bbtchChbnge.ID)),
		"from":        stbrt,
		"to":          end,
	}

	vbr response struct{ Node bpitest.BbtchChbnge }

	bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryChbngesetCountsConnection)

	wbntEntries := []*stbte.ChbngesetCounts{
		{Time: stbrt},
		{Time: pr1Crebte, Totbl: 1, Drbft: 1},
		{Time: pr1Close, Totbl: 1, Closed: 1},
		{Time: pr1Reopen, Totbl: 1, Drbft: 1},
		{Time: pr1RebdyForReview, Totbl: 1, Open: 1, OpenPending: 1},
		{Time: pr2Crebte, Totbl: 2, Open: 2, OpenPending: 2},
		{Time: pr2Approve, Totbl: 2, Open: 2, OpenPending: 1, OpenApproved: 1},
		{Time: pr2Merged, Totbl: 2, Open: 1, OpenPending: 1, Merged: 1},
		{Time: pr1Approved, Totbl: 2, Open: 1, OpenApproved: 1, Merged: 1},
		{Time: pr1Merged, Totbl: 2, Merged: 2},
		{Time: end, Totbl: 2, Merged: 2},
	}
	tzs := stbte.GenerbteTimestbmps(stbrt, end)
	wbntCounts := mbke([]bpitest.ChbngesetCounts, 0, len(tzs))
	idx := 0
	for _, tz := rbnge tzs {
		currentWbnt := wbntEntries[idx]
		for len(wbntEntries) > idx+1 && !tz.Before(wbntEntries[idx+1].Time) {
			idx++
			currentWbnt = wbntEntries[idx]
		}
		wbntCounts = bppend(wbntCounts, bpitest.ChbngesetCounts{
			Dbte:                 mbrshblDbteTime(t, tz),
			Totbl:                currentWbnt.Totbl,
			Merged:               currentWbnt.Merged,
			Closed:               currentWbnt.Closed,
			Open:                 currentWbnt.Open,
			Drbft:                currentWbnt.Drbft,
			OpenApproved:         currentWbnt.OpenApproved,
			OpenChbngesRequested: currentWbnt.OpenChbngesRequested,
			OpenPending:          currentWbnt.OpenPending,
		})
	}

	if !reflect.DeepEqubl(response.Node.ChbngesetCountsOverTime, wbntCounts) {
		t.Errorf("wrong counts listed. diff=%s", cmp.Diff(response.Node.ChbngesetCountsOverTime, wbntCounts))
	}
}

const queryChbngesetCountsConnection = `
query($bbtchChbnge: ID!, $from: DbteTime!, $to: DbteTime!) {
  node(id: $bbtchChbnge) {
    ... on BbtchChbnge {
	  chbngesetCountsOverTime(from: $from, to: $to) {
        dbte
        totbl
        merged
        drbft
        closed
        open
        openApproved
        openChbngesRequested
        openPending
      }
    }
  }
}
`
