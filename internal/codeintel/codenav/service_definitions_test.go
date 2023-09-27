pbckbge codenbv

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func TestDefinitions(t *testing.T) {
	// Set up mocks
	mockRepoStore := defbultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUplobdSvc := NewMockUplobdService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCbche, _ := NewHunkCbche(50)

	// Init service
	svc := newService(&observbtion.TestContext, mockRepoStore, mockLsifStore, mockUplobdSvc, mockGitserverClient)

	// Set up request stbte
	mockRequestStbte := RequestStbte{}
	mockRequestStbte.SetLocblCommitCbche(mockRepoStore, mockGitserverClient)

	mockRequestStbte.SetLocblGitTreeTrbnslbtor(mockGitserverClient, &sgtypes.Repo{}, mockCommit, mockPbth, hunkCbche)
	uplobds := []uplobdsshbred.Dump{
		{ID: 50, Commit: mockCommit, Root: "sub1/"},
		{ID: 51, Commit: mockCommit, Root: "sub2/"},
		{ID: 52, Commit: mockCommit, Root: "sub3/"},
		{ID: 53, Commit: mockCommit, Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetDefinitionLocbtionsFunc.PushReturn(locbtions, len(locbtions), nil)

	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 51,
			Commit:       mockCommit,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedLocbtions, err := svc.GetDefinitions(context.Bbckground(), mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying definitions: %s", err)
	}
	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: mockCommit, TbrgetRbnge: testRbnge1},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: mockCommit, TbrgetRbnge: testRbnge2},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: mockCommit, TbrgetRbnge: testRbnge3},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: mockCommit, TbrgetRbnge: testRbnge4},
		{Dump: uplobds[1], Pbth: "sub2/c.go", TbrgetCommit: mockCommit, TbrgetRbnge: testRbnge5},
	}

	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}

func TestDefinitionsWithSubRepoPermissions(t *testing.T) {
	// Set up mocks
	mockRepoStore := defbultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUplobdSvc := NewMockUplobdService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCbche, _ := NewHunkCbche(50)

	// Init service
	svc := newService(&observbtion.TestContext, mockRepoStore, mockLsifStore, mockUplobdSvc, mockGitserverClient)

	// Set up request stbte
	mockRequestStbte := RequestStbte{}
	mockRequestStbte.SetLocblCommitCbche(mockRepoStore, mockGitserverClient)
	mockRequestStbte.SetLocblGitTreeTrbnslbtor(mockGitserverClient, &sgtypes.Repo{}, mockCommit, mockPbth, hunkCbche)
	uplobds := []uplobdsshbred.Dump{
		{ID: 50, Commit: mockCommit, Root: "sub1/"},
		{ID: 51, Commit: mockCommit, Root: "sub2/"},
		{ID: 52, Commit: mockCommit, Root: "sub3/"},
		{ID: 53, Commit: mockCommit, Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)

	// Applying sub-repo permissions
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		if content.Pbth == "sub2/b.go" {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	mockRequestStbte.SetAuthChecker(checker)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetDefinitionLocbtionsFunc.PushReturn(locbtions, len(locbtions), nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 51,
			Commit:       "debdbeef",
		},
		Pbth:      "s1/mbin.go",
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedLocbtions, err := svc.GetDefinitions(ctx, mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying definitions: %s", err)
	}

	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge3},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}

func TestDefinitionsRemote(t *testing.T) {
	// Set up mocks
	mockRepoStore := defbultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUplobdSvc := NewMockUplobdService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCbche, _ := NewHunkCbche(50)

	// Init service
	svc := newService(&observbtion.TestContext, mockRepoStore, mockLsifStore, mockUplobdSvc, mockGitserverClient)

	// Set up request stbte
	mockRequestStbte := RequestStbte{}
	mockRequestStbte.SetLocblCommitCbche(mockRepoStore, mockGitserverClient)
	err := mockRequestStbte.SetLocblGitTreeTrbnslbtor(mockGitserverClient, &sgtypes.Repo{ID: 42}, mockCommit, mockPbth, hunkCbche)
	if err != nil {
		t.Fbtblf("unexpected error setting locbl git tree trbnslbtor: %s", err)
	}
	mockRequestStbte.GitTreeTrbnslbtor = mockedGitTreeTrbnslbtor()
	uplobds := []uplobdsshbred.Dump{
		{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		{ID: 53, Commit: "debdbeef", Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)

	dumps := []uplobdsshbred.Dump{
		{ID: 150, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 151, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 152, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 153, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(dumps, nil)

	// uplobd #150's commit no longer exists; bll others do
	mockGitserverClient.CommitsExistFunc.SetDefbultHook(func(ctx context.Context, _ buthz.SubRepoPermissionChecker, rcs []bpi.RepoCommit) (exists []bool, _ error) {
		for _, rc := rbnge rcs {
			exists = bppend(exists, rc.CommitID != "debdbeef1")
		}
		return
	})

	monikers := []precise.MonikerDbtb{
		{Kind: "import", Scheme: "tsc", Identifier: "pbdLeft", PbckbgeInformbtionID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pbd_left", PbckbgeInformbtionID: "52"},
		{Kind: "import", Scheme: "tsc", Identifier: "pbd-left", PbckbgeInformbtionID: "53"},
		{Kind: "import", Scheme: "tsc", Identifier: "left_pbd"},
	}
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[0]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[1]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[2]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[3]}}, nil)

	pbckbgeInformbtion1 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.1.0"}
	pbckbgeInformbtion2 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.2.0"}
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion1, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion2, true, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 151, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(locbtions, len(locbtions), nil)

	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}
	remoteUplobds := dumps
	bdjustedLocbtions, err := svc.GetDefinitions(context.Bbckground(), mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying definitions: %s", err)
	}

	xLocbtions := []shbred.UplobdLocbtion{
		{Dump: remoteUplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef2", TbrgetRbnge: testRbnge1},
		{Dump: remoteUplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef2", TbrgetRbnge: testRbnge2},
		{Dump: remoteUplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef2", TbrgetRbnge: testRbnge3},
		{Dump: remoteUplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef2", TbrgetRbnge: testRbnge4},
		{Dump: remoteUplobds[1], Pbth: "sub2/c.go", TbrgetCommit: "debdbeef2", TbrgetRbnge: testRbnge5},
	}

	if diff := cmp.Diff(xLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}

	if history := mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fbtblf("unexpected cbll count for dbstore.DefinitionDump. wbnt=%d hbve=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QublifiedMonikerDbtb{
			{MonikerDbtb: monikers[0], PbckbgeInformbtionDbtb: pbckbgeInformbtion1},
			{MonikerDbtb: monikers[2], PbckbgeInformbtionDbtb: pbckbgeInformbtion2},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocbtionsFunc.History(); len(history) != 1 {
		t.Fbtblf("unexpected cbll count for lsifstore.BulkMonikerResults. wbnt=%d hbve=%d", 1, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}

		expectedMonikers := []precise.MonikerDbtb{
			monikers[0],
			monikers[2],
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg3); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}
	}
}

func TestDefinitionsRemoteWithSubRepoPermissions(t *testing.T) {
	// Set up mocks
	mockRepoStore := defbultMockRepoStore()
	mockLsifStore := NewMockLsifStore()
	mockUplobdSvc := NewMockUplobdService()
	mockGitserverClient := gitserver.NewMockClient()
	hunkCbche, _ := NewHunkCbche(50)

	// Init service
	svc := newService(&observbtion.TestContext, mockRepoStore, mockLsifStore, mockUplobdSvc, mockGitserverClient)

	// Set up request stbte
	mockRequestStbte := RequestStbte{}
	mockRequestStbte.SetLocblCommitCbche(mockRepoStore, mockGitserverClient)
	mockRequestStbte.SetLocblGitTreeTrbnslbtor(mockGitserverClient, &sgtypes.Repo{ID: 42}, mockCommit, mockPbth, hunkCbche)
	uplobds := []uplobdsshbred.Dump{
		{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		{ID: 53, Commit: "debdbeef", Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)
	mockRequestStbte.GitTreeTrbnslbtor = mockedGitTreeTrbnslbtor()

	// Applying sub-repo permissions
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		if content.Pbth == "sub2/b.go" {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	mockRequestStbte.SetAuthChecker(checker)

	dumps := []uplobdsshbred.Dump{
		{ID: 150, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 151, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 152, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 153, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(dumps, nil)

	// uplobd #150's commit no longer exists; bll others do
	mockGitserverClient.CommitsExistFunc.SetDefbultHook(func(ctx context.Context, _ buthz.SubRepoPermissionChecker, rcs []bpi.RepoCommit) (exists []bool, _ error) {
		for _, rc := rbnge rcs {
			exists = bppend(exists, rc.CommitID != "debdbeef1")
		}
		return
	})

	monikers := []precise.MonikerDbtb{
		{Kind: "import", Scheme: "tsc", Identifier: "pbdLeft", PbckbgeInformbtionID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pbd_left", PbckbgeInformbtionID: "52"},
		{Kind: "import", Scheme: "tsc", Identifier: "pbd-left", PbckbgeInformbtionID: "53"},
		{Kind: "import", Scheme: "tsc", Identifier: "left_pbd"},
	}
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[0]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[1]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[2]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[3]}}, nil)

	pbckbgeInformbtion1 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.1.0"}
	pbckbgeInformbtion2 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.2.0"}
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion1, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion2, true, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 151, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(locbtions, 0, nil)
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(locbtions, len(locbtions), nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       "debdbeef",
		},
		Pbth:      "s1/mbin.go",
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedLocbtions, err := svc.GetDefinitions(ctx, mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying definitions: %s", err)
	}

	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: dumps[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef2", TbrgetRbnge: testRbnge2},
		{Dump: dumps[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef2", TbrgetRbnge: testRbnge4},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}

	if history := mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fbtblf("unexpected cbll count for dbstore.DefinitionDump. wbnt=%d hbve=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QublifiedMonikerDbtb{
			{MonikerDbtb: monikers[0], PbckbgeInformbtionDbtb: pbckbgeInformbtion1},
			{MonikerDbtb: monikers[2], PbckbgeInformbtionDbtb: pbckbgeInformbtion2},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocbtionsFunc.History(); len(history) != 1 {
		t.Fbtblf("unexpected cbll count for lsifstore.BulkMonikerResults. wbnt=%d hbve=%d", 1, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}

		expectedMonikers := []precise.MonikerDbtb{
			monikers[0],
			monikers[2],
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg3); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}
	}
}

func mockedGitTreeTrbnslbtor() GitTreeTrbnslbtor {
	mockPositionAdjuster := NewMockGitTreeTrbnslbtor()
	mockPositionAdjuster.GetTbrgetCommitPbthFromSourcePbthFunc.SetDefbultHook(func(ctx context.Context, commit string, pbth string, _ bool) (string, bool, error) {
		return commit, true, nil
	})
	mockPositionAdjuster.GetTbrgetCommitPositionFromSourcePositionFunc.SetDefbultHook(func(ctx context.Context, commit string, pos shbred.Position, _ bool) (string, shbred.Position, bool, error) {
		return commit, pos, true, nil
	})
	mockPositionAdjuster.GetTbrgetCommitRbngeFromSourceRbngeFunc.SetDefbultHook(func(ctx context.Context, commit string, pbth string, rx shbred.Rbnge, _ bool) (string, shbred.Rbnge, bool, error) {
		return commit, rx, true, nil
	})

	return mockPositionAdjuster
}
