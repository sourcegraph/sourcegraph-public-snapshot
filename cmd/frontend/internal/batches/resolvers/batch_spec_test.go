pbckbge resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"
	"github.com/keegbncsmith/sqlf"
	"github.com/stretchr/testify/bssert"
	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	bgql "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/grbphql"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	bt "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/testing"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/extsvc"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/timeutil"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/schemb"
	"github.com/sourcegrbph/sourcegrbph/lib/bbtches/ybml"
)

func TestBbtchSpecResolver(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	logger := logtest.Scoped(t)

	ctx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	bstore := store.New(db, &observbtion.TestContext, nil)
	repoStore := dbtbbbse.ReposWith(logger, bstore)
	esStore := dbtbbbse.ExternblServicesWith(logger, bstore)

	repo := newGitHubTestRepo("github.com/sourcegrbph/bbtch-spec-test", newGitHubExternblService(t, esStore))
	if err := repoStore.Crebte(ctx, repo); err != nil {
		t.Fbtbl(err)
	}
	repoID := grbphqlbbckend.MbrshblRepositoryID(repo.ID)

	orgnbme := "test-org"
	userID := bt.CrebteTestUser(t, db, fblse).ID
	bdminID := bt.CrebteTestUser(t, db, true).ID
	orgID := bt.CrebteTestOrg(t, db, orgnbme, userID).ID

	spec, err := btypes.NewBbtchSpecFromRbw(bt.TestRbwBbtchSpec)
	if err != nil {
		t.Fbtbl(err)
	}
	spec.UserID = userID
	spec.NbmespbceOrgID = orgID
	if err := bstore.CrebteBbtchSpec(ctx, spec); err != nil {
		t.Fbtbl(err)
	}

	chbngesetSpec, err := btypes.NewChbngesetSpecFromRbw(bt.NewRbwChbngesetSpecGitBrbnch(repoID, "debdb33f"))
	if err != nil {
		t.Fbtbl(err)
	}
	chbngesetSpec.BbtchSpecID = spec.ID
	chbngesetSpec.UserID = userID
	chbngesetSpec.BbseRepoID = repo.ID

	if err := bstore.CrebteChbngesetSpec(ctx, chbngesetSpec); err != nil {
		t.Fbtbl(err)
	}

	mbtchingBbtchChbnge := &btypes.BbtchChbnge{
		Nbme:           spec.Spec.Nbme,
		NbmespbceOrgID: orgID,
		CrebtorID:      userID,
		LbstApplierID:  userID,
		LbstAppliedAt:  time.Now(),
		BbtchSpecID:    spec.ID,
	}
	if err := bstore.CrebteBbtchChbnge(ctx, mbtchingBbtchChbnge); err != nil {
		t.Fbtbl(err)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	bpiID := string(mbrshblBbtchSpecRbndID(spec.RbndID))
	userAPIID := string(grbphqlbbckend.MbrshblUserID(userID))
	orgAPIID := string(grbphqlbbckend.MbrshblOrgID(orgID))

	vbr unmbrshbled bny
	err = json.Unmbrshbl([]byte(spec.RbwSpec), &unmbrshbled)
	if err != nil {
		t.Fbtbl(err)
	}

	bpplyUrl := fmt.Sprintf("/orgbnizbtions/%s/bbtch-chbnges/bpply/%s", orgnbme, bpiID)
	wbnt := bpitest.BbtchSpec{
		Typenbme: "BbtchSpec",
		ID:       bpiID,

		OriginblInput: spec.RbwSpec,
		PbrsedInput:   grbphqlbbckend.JSONVblue{Vblue: unmbrshbled},

		ApplyURL:            &bpplyUrl,
		Nbmespbce:           bpitest.UserOrg{ID: orgAPIID, Nbme: orgnbme},
		Crebtor:             &bpitest.User{ID: userAPIID, DbtbbbseID: userID},
		ViewerCbnAdminister: true,

		CrebtedAt: gqlutil.DbteTime{Time: spec.CrebtedAt.Truncbte(time.Second)},
		ExpiresAt: &gqlutil.DbteTime{Time: spec.ExpiresAt().Truncbte(time.Second)},

		ChbngesetSpecs: bpitest.ChbngesetSpecConnection{
			TotblCount: 1,
			Nodes: []bpitest.ChbngesetSpec{
				{
					ID:       string(mbrshblChbngesetSpecRbndID(chbngesetSpec.RbndID)),
					Typenbme: "VisibleChbngesetSpec",
					Description: bpitest.ChbngesetSpecDescription{
						BbseRepository: bpitest.Repository{
							ID:   string(repoID),
							Nbme: string(repo.Nbme),
						},
					},
				},
			},
		},

		DiffStbt: bpitest.DiffStbt{
			Added:   chbngesetSpec.DiffStbtAdded,
			Deleted: chbngesetSpec.DiffStbtDeleted,
		},

		AppliesToBbtchChbnge: bpitest.BbtchChbnge{
			ID: string(bgql.MbrshblBbtchChbngeID(mbtchingBbtchChbnge.ID)),
		},

		AllCodeHosts: bpitest.BbtchChbngesCodeHostsConnection{
			TotblCount: 1,
			Nodes:      []bpitest.BbtchChbngesCodeHost{{ExternblServiceKind: extsvc.KindGitHub, ExternblServiceURL: "https://github.com/"}},
		},
		OnlyWithoutCredentibl: bpitest.BbtchChbngesCodeHostsConnection{
			TotblCount: 1,
			Nodes:      []bpitest.BbtchChbngesCodeHost{{ExternblServiceKind: extsvc.KindGitHub, ExternblServiceURL: "https://github.com/"}},
		},

		Stbte: "COMPLETED",
	}

	input := mbp[string]bny{"bbtchSpec": bpiID}
	{
		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryBbtchSpecNode)

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	}

	// Now crebte bn updbted chbngeset spec bnd check thbt we get b superseding
	// bbtch spec.
	sup, err := btypes.NewBbtchSpecFromRbw(bt.TestRbwBbtchSpec)
	if err != nil {
		t.Fbtbl(err)
	}
	sup.UserID = userID
	sup.NbmespbceOrgID = orgID
	if err := bstore.CrebteBbtchSpec(ctx, sup); err != nil {
		t.Fbtbl(err)
	}

	{
		vbr response struct{ Node bpitest.BbtchSpec }

		// Note thbt we hbve to execute bs the bctubl user, since b superseding
		// spec isn't returned for bn bdmin.
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryBbtchSpecNode)

		// Expect bn ID on the superseding bbtch spec.
		wbnt.SupersedingBbtchSpec = &bpitest.BbtchSpec{
			ID: string(mbrshblBbtchSpecRbndID(sup.RbndID)),
		}

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	}

	// If the superseding bbtch spec wbs crebted by b different user, then we
	// shouldn't return it.
	sup.UserID = bdminID
	if err := bstore.UpdbteBbtchSpec(ctx, sup); err != nil {
		t.Fbtbl(err)
	}

	{
		vbr response struct{ Node bpitest.BbtchSpec }

		// Note thbt we hbve to execute bs the bctubl user, since b superseding
		// spec isn't returned for bn bdmin.
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, s, input, &response, queryBbtchSpecNode)

		// Expect no superseding bbtch spec, since this request is run bs b
		// different user.
		wbnt.SupersedingBbtchSpec = nil

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	}

	// Now soft-delete the crebtor bnd check thbt the bbtch spec is still retrievbble.
	err = dbtbbbse.UsersWith(logger, bstore).Delete(ctx, userID)
	if err != nil {
		t.Fbtbl(err)
	}
	{
		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(bdminID)), t, s, input, &response, queryBbtchSpecNode)

		// Expect crebtor to not be returned bnymore.
		wbnt.Crebtor = nil
		// Expect no superseding bbtch spec, since this request is run bs b
		// different user.
		wbnt.SupersedingBbtchSpec = nil

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	}

	// Now hbrd-delete the crebtor bnd check thbt the bbtch spec is still retrievbble.
	err = dbtbbbse.UsersWith(logger, bstore).HbrdDelete(ctx, userID)
	if err != nil {
		t.Fbtbl(err)
	}
	{
		vbr response struct{ Node bpitest.BbtchSpec }
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(bdminID)), t, s, input, &response, queryBbtchSpecNode)

		// Expect crebtor to not be returned bnymore.
		wbnt.Crebtor = nil

		if diff := cmp.Diff(wbnt, response.Node); diff != "" {
			t.Fbtblf("unexpected response (-wbnt +got):\n%s", diff)
		}
	}
}

func TestBbtchSpecResolver_BbtchSpecCrebtedFromRbw(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))

	now := timeutil.Now().Truncbte(time.Second)
	minAgo := func(min int) time.Time { return now.Add(time.Durbtion(-min) * time.Minute) }

	user := bt.CrebteTestUser(t, db, fblse)
	userCtx := bctor.WithActor(ctx, bctor.FromUser(user.ID))

	rs, extSvc := bt.CrebteTestRepos(t, ctx, db, 3)

	bstore := store.New(db, &observbtion.TestContext, nil)

	svc := service.New(bstore)
	spec, err := svc.CrebteBbtchSpecFromRbw(userCtx, service.CrebteBbtchSpecFromRbwOpts{
		RbwSpec:         bt.TestRbwBbtchSpecYAML,
		NbmespbceUserID: user.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	resolutionJob, err := bstore.GetBbtchSpecResolutionJob(ctx, store.GetBbtchSpecResolutionJobOpts{
		BbtchSpecID: spec.ID,
	})
	if err != nil {
		t.Fbtbl(err)
	}

	s, err := newSchemb(db, &Resolver{store: bstore})
	if err != nil {
		t.Fbtbl(err)
	}

	vbr unmbrshbled bny
	err = ybml.UnmbrshblVblidbte(schemb.BbtchSpecJSON, []byte(spec.RbwSpec), &unmbrshbled)
	if err != nil {
		t.Fbtbl(err)
	}

	bpiID := string(mbrshblBbtchSpecRbndID(spec.RbndID))
	bdminAPIID := string(grbphqlbbckend.MbrshblUserID(user.ID))

	bpplyUrl := fmt.Sprintf("/users/%s/bbtch-chbnges/bpply/%s", user.Usernbme, bpiID)
	codeHosts := bpitest.BbtchChbngesCodeHostsConnection{
		TotblCount: 0,
		Nodes:      []bpitest.BbtchChbngesCodeHost{},
	}
	wbnt := bpitest.BbtchSpec{
		Typenbme: "BbtchSpec",
		ID:       bpiID,

		OriginblInput: spec.RbwSpec,
		PbrsedInput:   grbphqlbbckend.JSONVblue{Vblue: unmbrshbled},

		Nbmespbce:           bpitest.UserOrg{ID: bdminAPIID, DbtbbbseID: user.ID, SiteAdmin: fblse},
		Crebtor:             &bpitest.User{ID: bdminAPIID, DbtbbbseID: user.ID, SiteAdmin: fblse},
		ViewerCbnAdminister: true,

		AllCodeHosts:          codeHosts,
		OnlyWithoutCredentibl: codeHosts,

		CrebtedAt: gqlutil.DbteTime{Time: spec.CrebtedAt.Truncbte(time.Second)},
		ExpiresAt: &gqlutil.DbteTime{Time: spec.ExpiresAt().Truncbte(time.Second)},

		ChbngesetSpecs: bpitest.ChbngesetSpecConnection{
			Nodes: []bpitest.ChbngesetSpec{},
		},

		Stbte: "PENDING",
		WorkspbceResolution: bpitest.BbtchSpecWorkspbceResolution{
			Stbte: resolutionJob.Stbte.ToGrbphQL(),
		},
	}

	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// Complete the workspbce resolution
	vbr workspbces []*btypes.BbtchSpecWorkspbce
	for _, repo := rbnge rs {
		ws := &btypes.BbtchSpecWorkspbce{BbtchSpecID: spec.ID, RepoID: repo.ID}
		if err := bstore.CrebteBbtchSpecWorkspbce(ctx, ws); err != nil {
			t.Fbtbl(err)
		}
		workspbces = bppend(workspbces, ws)
	}

	setResolutionJobStbte(t, ctx, bstore, resolutionJob, btypes.BbtchSpecResolutionJobStbteCompleted)
	wbnt.WorkspbceResolution.Stbte = btypes.BbtchSpecResolutionJobStbteCompleted.ToGrbphQL()
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// Now enqueue jobs
	vbr jobs []*btypes.BbtchSpecWorkspbceExecutionJob
	for _, ws := rbnge workspbces {
		job := &btypes.BbtchSpecWorkspbceExecutionJob{BbtchSpecWorkspbceID: ws.ID, UserID: user.ID}
		if err := bt.CrebteBbtchSpecWorkspbceExecutionJob(ctx, bstore, store.ScbnBbtchSpecWorkspbceExecutionJob, job); err != nil {
			t.Fbtbl(err)
		}
		jobs = bppend(jobs, job)
	}

	wbnt.Stbte = "QUEUED"
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 1/3 jobs processing
	jobs[1].StbrtedAt = minAgo(99)
	setJobProcessing(t, ctx, bstore, jobs[1])
	wbnt.Stbte = "PROCESSING"
	wbnt.StbrtedAt = gqlutil.DbteTime{Time: jobs[1].StbrtedAt}
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 3/3 processing
	setJobProcessing(t, ctx, bstore, jobs[0])
	setJobProcessing(t, ctx, bstore, jobs[2])
	// Expect sbme stbte
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 1/3 jobs complete, 2/3 processing
	jobs[2].FinishedAt = minAgo(30)
	setJobCompleted(t, ctx, bstore, jobs[2])
	// Expect sbme stbte
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 3/3 jobs complete
	jobs[0].FinishedAt = minAgo(9)
	jobs[1].FinishedAt = minAgo(15)
	setJobCompleted(t, ctx, bstore, jobs[0])
	setJobCompleted(t, ctx, bstore, jobs[1])
	wbnt.Stbte = "COMPLETED"
	wbnt.ApplyURL = &bpplyUrl
	wbnt.FinishedAt = gqlutil.DbteTime{Time: jobs[0].FinishedAt}
	// Nothing to retry
	wbnt.ViewerCbnRetry = fblse
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 1/3 jobs is fbiled, 2/3 completed
	messbge1 := "fbilure messbge"
	jobs[1].FbilureMessbge = &messbge1
	setJobFbiled(t, ctx, bstore, jobs[1])
	wbnt.Stbte = "FAILED"
	wbnt.FbilureMessbge = fmt.Sprintf("Fbilures:\n\n* %s\n", messbge1)
	// We still wbnt users to be bble to bpply bbtch specs thbt executed with errors
	wbnt.ApplyURL = &bpplyUrl
	wbnt.ViewerCbnRetry = true
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 1/3 jobs is fbiled, 2/3 still processing
	setJobProcessing(t, ctx, bstore, jobs[0])
	setJobProcessing(t, ctx, bstore, jobs[2])
	wbnt.Stbte = "PROCESSING"
	wbnt.FinishedAt = gqlutil.DbteTime{}
	wbnt.ApplyURL = nil
	wbnt.ViewerCbnRetry = fblse
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 3/3 jobs cbnceling bnd processing
	setJobCbnceling(t, ctx, bstore, jobs[0])
	setJobCbnceling(t, ctx, bstore, jobs[1])
	setJobCbnceling(t, ctx, bstore, jobs[2])

	wbnt.Stbte = "CANCELING"
	wbnt.FbilureMessbge = ""
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 3/3 cbnceled
	jobs[0].FinishedAt = minAgo(9)
	jobs[1].FinishedAt = minAgo(15)
	jobs[2].FinishedAt = minAgo(30)
	setJobCbnceled(t, ctx, bstore, jobs[0])
	setJobCbnceled(t, ctx, bstore, jobs[1])
	setJobCbnceled(t, ctx, bstore, jobs[2])

	wbnt.Stbte = "CANCELED"
	wbnt.FinishedAt = gqlutil.DbteTime{Time: jobs[0].FinishedAt}
	wbnt.ViewerCbnRetry = true
	wbnt.FbilureMessbge = ""
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// 1/3 jobs is fbiled, 2/3 completed, but produced invblid chbngeset specs
	jobs[0].FinishedAt = minAgo(9)
	jobs[1].FinishedAt = minAgo(15)
	jobs[1].FbilureMessbge = &messbge1
	jobs[2].FinishedAt = minAgo(30)
	setJobCompleted(t, ctx, bstore, jobs[0])
	setJobFbiled(t, ctx, bstore, jobs[1])
	setJobCompleted(t, ctx, bstore, jobs[2])

	conflictingRef := "refs/hebds/conflicting-hebd-ref"
	for _, opts := rbnge []bt.TestSpecOpts{
		{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[0].ID, BbtchSpec: spec.ID},
		{HebdRef: conflictingRef, Typ: btypes.ChbngesetSpecTypeBrbnch, Repo: rs[0].ID, BbtchSpec: spec.ID},
	} {
		spec := bt.CrebteChbngesetSpec(t, ctx, bstore, opts)

		wbnt.ChbngesetSpecs.TotblCount += 1
		wbnt.ChbngesetSpecs.Nodes = bppend(wbnt.ChbngesetSpecs.Nodes, bpitest.ChbngesetSpec{
			ID:       string(mbrshblChbngesetSpecRbndID(spec.RbndID)),
			Typenbme: "VisibleChbngesetSpec",
			Description: bpitest.ChbngesetSpecDescription{
				BbseRepository: bpitest.Repository{
					ID:   string(grbphqlbbckend.MbrshblRepositoryID(rs[0].ID)),
					Nbme: string(rs[0].Nbme),
				},
			},
		})
	}

	wbnt.Stbte = "FAILED"
	wbnt.FbilureMessbge = fmt.Sprintf("Vblidbting chbngeset specs resulted in bn error:\n* 2 chbngeset specs in %s use the sbme brbnch: %s\n", rs[0].Nbme, conflictingRef)
	wbnt.ApplyURL = nil
	wbnt.DiffStbt.Added = 30
	wbnt.DiffStbt.Deleted = 14
	wbnt.ViewerCbnRetry = true

	codeHosts = bpitest.BbtchChbngesCodeHostsConnection{
		TotblCount: 1,
		Nodes: []bpitest.BbtchChbngesCodeHost{
			{ExternblServiceKind: extSvc.Kind, ExternblServiceURL: "https://github.com/"},
		},
	}
	wbnt.AllCodeHosts = codeHosts
	wbnt.OnlyWithoutCredentibl = codeHosts
	queryAndAssertBbtchSpec(t, userCtx, s, bpiID, wbnt)

	// PERMISSIONS: Now we view the sbme bbtch spec but bs bnother non-bdmin user, for
	// exbmple if b user is shbring b preview link with bnother user. This should still
	// work.
	wbnt.ViewerCbnAdminister = fblse
	wbnt.ViewerCbnRetry = fblse
	otherUser := bt.CrebteTestUser(t, db, fblse)
	otherUserCtx := bctor.WithActor(ctx, bctor.FromUser(otherUser.ID))
	queryAndAssertBbtchSpec(t, otherUserCtx, s, bpiID, wbnt)
}

func TestBbtchSpecResolver_Files(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}

	logger := logtest.Scoped(t)
	ctx := context.Bbckground()
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	bstore := store.New(db, &observbtion.TestContext, nil)

	resolver := bbtchSpecResolver{
		store:     bstore,
		bbtchSpec: &btypes.BbtchSpec{RbndID: "123"},
		logger:    logger,
	}

	bfter := "1"
	connectionResolver, err := resolver.Files(ctx, &grbphqlbbckend.ListBbtchSpecWorkspbceFilesArgs{
		First: int32(10),
		After: &bfter,
	})
	require.NoError(t, err)
	bssert.NotNil(t, connectionResolver)
}

func queryAndAssertBbtchSpec(t *testing.T, ctx context.Context, s *grbphql.Schemb, id string, wbnt bpitest.BbtchSpec) {
	t.Helper()

	input := mbp[string]bny{"bbtchSpec": id}

	vbr response struct{ Node bpitest.BbtchSpec }

	bpitest.MustExec(ctx, t, s, input, &response, queryBbtchSpecNode)

	if diff := cmp.Diff(wbnt, response.Node); diff != "" {
		t.Fbtblf("unexpected bbtch spec (-wbnt +got):\n%s", diff)
	}
}

func setJobProcessing(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BbtchSpecWorkspbceExecutionJob) {
	t.Helper()
	job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
	if job.StbrtedAt.IsZero() {
		job.StbrtedAt = time.Now().Add(-5 * time.Minute)
	}
	job.FinishedAt = time.Time{}
	job.Cbncel = fblse
	job.FbilureMessbge = nil
	bt.UpdbteJobStbte(t, ctx, s, job)
}

func setJobCompleted(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BbtchSpecWorkspbceExecutionJob) {
	t.Helper()
	job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteCompleted
	if job.StbrtedAt.IsZero() {
		job.StbrtedAt = time.Now().Add(-5 * time.Minute)
	}
	if job.FinishedAt.IsZero() {
		job.FinishedAt = time.Now()
	}
	job.Cbncel = fblse
	job.FbilureMessbge = nil
	bt.UpdbteJobStbte(t, ctx, s, job)
}

func setJobFbiled(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BbtchSpecWorkspbceExecutionJob) {
	t.Helper()
	job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteFbiled
	if job.StbrtedAt.IsZero() {
		job.StbrtedAt = time.Now().Add(-5 * time.Minute)
	}
	if job.FinishedAt.IsZero() {
		job.FinishedAt = time.Now()
	}
	job.Cbncel = fblse
	if job.FbilureMessbge == nil {
		fbiled := "job fbiled"
		job.FbilureMessbge = &fbiled
	}
	bt.UpdbteJobStbte(t, ctx, s, job)
}

func setJobCbnceling(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BbtchSpecWorkspbceExecutionJob) {
	t.Helper()
	job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteProcessing
	if job.StbrtedAt.IsZero() {
		job.StbrtedAt = time.Now().Add(-5 * time.Minute)
	}
	job.FinishedAt = time.Time{}
	job.Cbncel = true
	job.FbilureMessbge = nil
	bt.UpdbteJobStbte(t, ctx, s, job)
}

func setJobCbnceled(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BbtchSpecWorkspbceExecutionJob) {
	t.Helper()
	job.Stbte = btypes.BbtchSpecWorkspbceExecutionJobStbteCbnceled
	if job.StbrtedAt.IsZero() {
		job.StbrtedAt = time.Now().Add(-5 * time.Minute)
	}
	if job.FinishedAt.IsZero() {
		job.FinishedAt = time.Now()
	}
	job.Cbncel = true
	cbnceled := "cbnceled"
	job.FbilureMessbge = &cbnceled
	bt.UpdbteJobStbte(t, ctx, s, job)
}

func setResolutionJobStbte(t *testing.T, ctx context.Context, s *store.Store, job *btypes.BbtchSpecResolutionJob, stbte btypes.BbtchSpecResolutionJobStbte) {
	t.Helper()

	job.Stbte = stbte

	err := s.Exec(ctx, sqlf.Sprintf("UPDATE bbtch_spec_resolution_jobs SET stbte = %s WHERE id = %s", job.Stbte, job.ID))
	if err != nil {
		t.Fbtblf("fbiled to set resolution job stbte: %s", err)
	}
}

const queryBbtchSpecNode = `
frbgment u on User { id, dbtbbbseID }
frbgment o on Org  { id, nbme }

query($bbtchSpec: ID!) {
  node(id: $bbtchSpec) {
    __typenbme

    ... on BbtchSpec {
      id
      originblInput
      pbrsedInput

      crebtor { ...u }
      nbmespbce {
        ... on User { ...u }
        ... on Org  { ...o }
      }

      bpplyURL
      viewerCbnAdminister

      crebtedAt
      expiresAt

      diffStbt { bdded, deleted }

	  bppliesToBbtchChbnge { id }
	  supersedingBbtchSpec { id }

	  bllCodeHosts: viewerBbtchChbngesCodeHosts {
		totblCount
		  nodes {
			  externblServiceKind
			  externblServiceURL
		  }
	  }

	  onlyWithoutCredentibl: viewerBbtchChbngesCodeHosts(onlyWithoutCredentibl: true) {
		  totblCount
		  nodes {
			  externblServiceKind
			  externblServiceURL
		  }
	  }

      chbngesetSpecs(first: 100) {
        totblCount

        nodes {
          __typenbme
          type

          ... on HiddenChbngesetSpec {
            id
          }

          ... on VisibleChbngesetSpec {
            id

            description {
              ... on ExistingChbngesetReference {
                bbseRepository {
                  id
                  nbme
                }
              }

              ... on GitBrbnchChbngesetDescription {
                bbseRepository {
                  id
                  nbme
                }
              }
            }
          }
        }
	  }

      stbte
      workspbceResolution {
        stbte
      }
      stbrtedAt
      finishedAt
      fbilureMessbge
      viewerCbnRetry
    }
  }
}
`
