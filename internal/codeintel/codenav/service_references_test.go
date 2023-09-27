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

vbr (
	testRbnge1 = shbred.Rbnge{Stbrt: shbred.Position{Line: 11, Chbrbcter: 21}, End: shbred.Position{Line: 31, Chbrbcter: 41}}
	testRbnge2 = shbred.Rbnge{Stbrt: shbred.Position{Line: 12, Chbrbcter: 22}, End: shbred.Position{Line: 32, Chbrbcter: 42}}
	testRbnge3 = shbred.Rbnge{Stbrt: shbred.Position{Line: 13, Chbrbcter: 23}, End: shbred.Position{Line: 33, Chbrbcter: 43}}
	testRbnge4 = shbred.Rbnge{Stbrt: shbred.Position{Line: 14, Chbrbcter: 24}, End: shbred.Position{Line: 34, Chbrbcter: 44}}
	testRbnge5 = shbred.Rbnge{Stbrt: shbred.Position{Line: 15, Chbrbcter: 25}, End: shbred.Position{Line: 35, Chbrbcter: 45}}
	testRbnge6 = shbred.Rbnge{Stbrt: shbred.Position{Line: 16, Chbrbcter: 26}, End: shbred.Position{Line: 36, Chbrbcter: 46}}

	mockPbth   = "s1/mbin.go"
	mockCommit = "debdbeef"
)

func TestReferences(t *testing.T) {
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
		{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		{ID: 53, Commit: "debdbeef", Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)

	// Empty result set (prevents nil pointer bs scbnner is blwbys non-nil)
	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[4:], 1, nil)

	mockCursor := ReferencesCursor{Phbse: "locbl"}
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedLocbtions, _, err := svc.GetReferences(context.Bbckground(), mockRequest, mockRequestStbte, mockCursor)
	if err != nil {
		t.Fbtblf("unexpected error querying references: %s", err)
	}

	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge2},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge3},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge4},
		{Dump: uplobds[1], Pbth: "sub2/c.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge5},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}

func TestReferencesWithSubRepoPermissions(t *testing.T) {
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
		{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		{ID: 53, Commit: "debdbeef", Root: "sub4/"},
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

	// Empty result set (prevents nil pointer bs scbnner is blwbys non-nil)
	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[4:], 1, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	mockCursor := ReferencesCursor{Phbse: "locbl"}
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}

	bdjustedLocbtions, _, err := svc.GetReferences(ctx, mockRequest, mockRequestStbte, mockCursor)
	if err != nil {
		t.Fbtblf("unexpected error querying references: %s", err)
	}
	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge3},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}

func TestReferencesRemote(t *testing.T) {
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
		{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		{ID: 53, Commit: "debdbeef", Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)

	definitionUplobds := []uplobdsshbred.Dump{
		{ID: 150, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 151, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 152, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 153, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(definitionUplobds, nil)

	referenceUplobds := []uplobdsshbred.Dump{
		{ID: 250, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 251, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 252, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 253, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsByIDsFunc.PushReturn(nil, nil) // empty
	mockUplobdSvc.GetDumpsByIDsFunc.PushReturn(referenceUplobds[:2], nil)
	mockUplobdSvc.GetDumpsByIDsFunc.PushReturn(referenceUplobds[2:], nil)

	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{250, 251}, 0, 4, nil)
	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{252, 253}, 0, 2, nil)

	// uplobd #150/#250's commits no longer exists; bll others do
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
	pbckbgeInformbtion3 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.3.0"}
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion1, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion2, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion3, true, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[4:5], 1, nil)

	monikerLocbtions := []shbred.Locbtion{
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 53, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[0:1], 1, nil) // defs
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[1:2], 1, nil) // refs bbtch 1
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[2:], 3, nil)  // refs bbtch 2

	// uplobds := []dbstore.Dump{
	// 	{ID: 50, Commit: "debdbeef", Root: "sub1/"},
	// 	{ID: 51, Commit: "debdbeef", Root: "sub2/"},
	// 	{ID: 52, Commit: "debdbeef", Root: "sub3/"},
	// 	{ID: 53, Commit: "debdbeef", Root: "sub4/"},
	// }
	// resolver.SetUplobdsDbtbLobder(uplobds)

	mockCursor := ReferencesCursor{Phbse: "locbl"}
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedLocbtions, _, err := svc.GetReferences(context.Bbckground(), mockRequest, mockRequestStbte, mockCursor)
	if err != nil {
		t.Fbtblf("unexpected error querying references: %s", err)
	}

	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge2},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge3},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge4},
		{Dump: uplobds[1], Pbth: "sub2/c.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge5},
		{Dump: uplobds[3], Pbth: "sub4/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
		{Dump: uplobds[3], Pbth: "sub4/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge2},
		{Dump: uplobds[3], Pbth: "sub4/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge3},
		{Dump: uplobds[3], Pbth: "sub4/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge4},
		{Dump: uplobds[3], Pbth: "sub4/c.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge5},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}

	if history := mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fbtblf("unexpected cbll count for dbstore.DefinitionDump. wbnt=%d hbve=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QublifiedMonikerDbtb{
			{MonikerDbtb: monikers[0], PbckbgeInformbtionDbtb: pbckbgeInformbtion1},
			{MonikerDbtb: monikers[1], PbckbgeInformbtionDbtb: pbckbgeInformbtion2},
			{MonikerDbtb: monikers[2], PbckbgeInformbtionDbtb: pbckbgeInformbtion3},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocbtionsFunc.History(); len(history) != 3 {
		t.Fbtblf("unexpected cbll count for lsifstore.BulkMonikerResults. wbnt=%d hbve=%d", 3, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}

		expectedMonikers := []precise.MonikerDbtb{
			monikers[0],
			monikers[1],
			monikers[2],
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{251}, history[1].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}
		if diff := cmp.Diff(expectedMonikers, history[1].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{252, 253}, history[2].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}
		if diff := cmp.Diff(expectedMonikers, history[2].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}
	}
}

func TestReferencesRemoteWithSubRepoPermissions(t *testing.T) {
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
		{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		{ID: 53, Commit: "debdbeef", Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)

	// Applying sub-repo permissions
	checker := buthz.NewMockSubRepoPermissionChecker()
	checker.EnbbledFunc.SetDefbultHook(func() bool {
		return true
	})
	checker.PermissionsFunc.SetDefbultHook(func(ctx context.Context, i int32, content buthz.RepoContent) (buthz.Perms, error) {
		if content.Pbth == "sub2/b.go" || content.Pbth == "sub4/b.go" {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	mockRequestStbte.SetAuthChecker(checker)

	definitionUplobds := []uplobdsshbred.Dump{
		{ID: 150, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 151, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 152, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 153, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(definitionUplobds, nil)

	referenceUplobds := []uplobdsshbred.Dump{
		{ID: 250, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 251, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 252, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 253, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsByIDsFunc.PushReturn(nil, nil) // empty
	mockUplobdSvc.GetDumpsByIDsFunc.PushReturn(referenceUplobds[:2], nil)
	mockUplobdSvc.GetDumpsByIDsFunc.PushReturn(referenceUplobds[2:], nil)

	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{250, 251}, 0, 4, nil)
	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{252, 253}, 0, 2, nil)

	// uplobd #150/#250's commits no longer exists; bll others do
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
	pbckbgeInformbtion3 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.3.0"}
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion1, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion2, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion3, true, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetReferenceLocbtionsFunc.PushReturn(locbtions[4:5], 1, nil)

	monikerLocbtions := []shbred.Locbtion{
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 53, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[0:1], 1, nil) // defs
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[1:2], 1, nil) // refs bbtch 1
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[2:], 3, nil)  // refs bbtch 2

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	mockCursor := ReferencesCursor{Phbse: "locbl"}
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        50,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedLocbtions, _, err := svc.GetReferences(ctx, mockRequest, mockRequestStbte, mockCursor)
	if err != nil {
		t.Fbtblf("unexpected error querying references: %s", err)
	}

	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge2},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge4},
		{Dump: uplobds[3], Pbth: "sub4/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge2},
		{Dump: uplobds[3], Pbth: "sub4/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge4},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}

	if history := mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.History(); len(history) != 1 {
		t.Fbtblf("unexpected cbll count for dbstore.DefinitionDump. wbnt=%d hbve=%d", 1, len(history))
	} else {
		expectedMonikers := []precise.QublifiedMonikerDbtb{
			{MonikerDbtb: monikers[0], PbckbgeInformbtionDbtb: pbckbgeInformbtion1},
			{MonikerDbtb: monikers[1], PbckbgeInformbtionDbtb: pbckbgeInformbtion2},
			{MonikerDbtb: monikers[2], PbckbgeInformbtionDbtb: pbckbgeInformbtion3},
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}
	}

	if history := mockLsifStore.GetBulkMonikerLocbtionsFunc.History(); len(history) != 3 {
		t.Fbtblf("unexpected cbll count for mockSvc.GetBulkMonikerLocbtionsFunc. wbnt=%d hbve=%d", 3, len(history))
	} else {
		if diff := cmp.Diff([]int{151, 152, 153}, history[0].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}

		expectedMonikers := []precise.MonikerDbtb{
			monikers[0],
			monikers[1],
			monikers[2],
		}
		if diff := cmp.Diff(expectedMonikers, history[0].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{251}, history[1].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}
		if diff := cmp.Diff(expectedMonikers, history[1].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}

		if diff := cmp.Diff([]int{252, 253}, history[2].Arg2); diff != "" {
			t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
		}
		if diff := cmp.Diff(expectedMonikers, history[2].Arg3); diff != "" {
			t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
		}
	}
}
