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

func TestImplementbtions(t *testing.T) {
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

	// Empty result set (prevents nil pointer bs scbnner is blwbys non-nil)
	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[4:], 1, nil)

	uplobds := []uplobdsshbred.Dump{
		{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		{ID: 53, Commit: "debdbeef", Root: "sub4/"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)
	mockCursor := ImplementbtionsCursor{Phbse: "locbl"}
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 51,
			Commit:       "debdbeef",
			Limit:        50,
		},
		Pbth:      "s1/mbin.go",
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedLocbtions, _, err := svc.GetImplementbtions(context.Bbckground(), mockRequest, mockRequestStbte, mockCursor)
	if err != nil {
		t.Fbtblf("unexpected error querying implementbtions: %s", err)
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

func TestImplementbtionsWithSubRepoPermissions(t *testing.T) {
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

	// Empty result set (prevents nil pointer bs scbnner is blwbys non-nil)
	mockUplobdSvc.GetUplobdIDsWithReferencesFunc.PushReturn([]int{}, 0, 0, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[4:], 1, nil)

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

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	mockCursor := ImplementbtionsCursor{Phbse: "locbl"}
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
	bdjustedLocbtions, _, err := svc.GetImplementbtions(ctx, mockRequest, mockRequestStbte, mockCursor)
	if err != nil {
		t.Fbtblf("unexpected error querying implementbtions: %s", err)
	}

	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge3},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}

func TestImplementbtionsRemote(t *testing.T) {
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

	remoteUplobds := []uplobdsshbred.Dump{
		{ID: 150, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 151, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 152, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 153, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(remoteUplobds, nil)

	referenceUplobds := []uplobdsshbred.Dump{
		{ID: 250, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 251, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 252, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 253, Commit: "debdbeef4", Root: "sub4/"},
	}
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
		{Kind: "implementbtion", Scheme: "tsc", Identifier: "pbdLeft", PbckbgeInformbtionID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pbd_left", PbckbgeInformbtionID: "52"},
	}
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[0]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[1]}}, nil)

	pbckbgeInformbtion1 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.1.0"}
	pbckbgeInformbtion2 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.2.0"}
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion1, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion2, true, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[4:5], 1, nil)

	monikerLocbtions := []shbred.Locbtion{
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 53, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[0:1], 1, nil) // defs
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[1:2], 1, nil) // impls bbtch 1
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[2:], 3, nil)  // impls bbtch 2

	mockCursor := ImplementbtionsCursor{Phbse: "locbl"}
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
	bdjustedLocbtions, _, err := svc.GetImplementbtions(context.Bbckground(), mockRequest, mockRequestStbte, mockCursor)
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
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}

func TestImplementbtionsRemoteWithSubRepoPermissions(t *testing.T) {
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
		{Kind: "implementbtion", Scheme: "tsc", Identifier: "pbdLeft", PbckbgeInformbtionID: "51"},
		{Kind: "export", Scheme: "tsc", Identifier: "pbd_left", PbckbgeInformbtionID: "52"},
	}
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[0]}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{}}, nil)
	mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[1]}}, nil)

	pbckbgeInformbtion1 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.1.0"}
	pbckbgeInformbtion2 := precise.PbckbgeInformbtionDbtb{Nbme: "leftpbd", Version: "0.2.0"}
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion1, true, nil)
	mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion2, true, nil)

	locbtions := []shbred.Locbtion{
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[:1], 1, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[1:4], 3, nil)
	mockLsifStore.GetImplementbtionLocbtionsFunc.PushReturn(locbtions[4:5], 1, nil)

	monikerLocbtions := []shbred.Locbtion{
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge1},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge2},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge3},
		{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge4},
		{DumpID: 53, Pbth: "c.go", Rbnge: testRbnge5},
	}
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[0:1], 1, nil) // defs
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[1:2], 1, nil) // impls bbtch 1
	mockLsifStore.GetBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[2:], 3, nil)  // impls bbtch 2

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	mockCursor := ImplementbtionsCursor{Phbse: "locbl"}
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
	bdjustedLocbtions, _, err := svc.GetImplementbtions(ctx, mockRequest, mockRequestStbte, mockCursor)
	if err != nil {
		t.Fbtblf("unexpected error querying references: %s", err)
	}

	expectedLocbtions := []shbred.UplobdLocbtion{
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
		{Dump: uplobds[1], Pbth: "sub2/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge3},
		{Dump: uplobds[3], Pbth: "sub4/b.go", TbrgetCommit: "debdbeef", TbrgetRbnge: testRbnge1},
	}
	if diff := cmp.Diff(expectedLocbtions, bdjustedLocbtions); diff != "" {
		t.Errorf("unexpected locbtions (-wbnt +got):\n%s", diff)
	}
}
