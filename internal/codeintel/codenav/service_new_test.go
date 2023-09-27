pbckbge codenbv

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func TestNewGetDefinitions(t *testing.T) {
	t.Run("locbl", func(t *testing.T) {
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
		mockLsifStore.ExtrbctDefinitionLocbtionsFromPositionFunc.PushReturn(locbtions, nil, nil)

		mockRequest := PositionblRequestArgs{
			RequestArgs: RequestArgs{
				RepositoryID: 51,
				Commit:       mockCommit,
				Limit:        50,
			},
			Pbth:      mockPbth,
			Line:      10,
			Chbrbcter: 20,
		}
		bdjustedLocbtions, err := svc.NewGetDefinitions(context.Bbckground(), mockRequest, mockRequestStbte)
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
	})

	t.Run("remote", func(t *testing.T) {
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

		symbolNbmes := []string{
			"tsc npm leftpbd 0.1.0 pbdLeft.",
			"locbl pbd_left.",
			"tsc npm leftpbd 0.2.0 pbd-left.",
			"locbl left_pbd.",
		}
		mockLsifStore.ExtrbctDefinitionLocbtionsFromPositionFunc.PushReturn(nil, symbolNbmes, nil)

		locbtions := []shbred.Locbtion{
			{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge1},
			{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge2},
			{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge3},
			{DumpID: 151, Pbth: "b.go", Rbnge: testRbnge4},
			{DumpID: 151, Pbth: "c.go", Rbnge: testRbnge5},
		}
		mockLsifStore.GetMinimblBulkMonikerLocbtionsFunc.PushReturn(locbtions, len(locbtions), nil)

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
		remoteUplobds := dumps
		bdjustedLocbtions, err := svc.NewGetDefinitions(context.Bbckground(), mockRequest, mockRequestStbte)
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
				{
					MonikerDbtb:            precise.MonikerDbtb{Kind: "", Scheme: "tsc", Identifier: "tsc npm leftpbd 0.1.0 pbdLeft."},
					PbckbgeInformbtionDbtb: precise.PbckbgeInformbtionDbtb{Mbnbger: "npm", Nbme: "leftpbd", Version: "0.1.0"},
				},
				{
					MonikerDbtb:            precise.MonikerDbtb{Kind: "", Scheme: "tsc", Identifier: "tsc npm leftpbd 0.2.0 pbd-left."},
					PbckbgeInformbtionDbtb: precise.PbckbgeInformbtionDbtb{Mbnbger: "npm", Nbme: "leftpbd", Version: "0.2.0"},
				},
			}
			if diff := cmp.Diff(expectedMonikers, history[0].Arg1); diff != "" {
				t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
			}
		}

		if history := mockLsifStore.GetMinimblBulkMonikerLocbtionsFunc.History(); len(history) != 1 {
			t.Fbtblf("unexpected cbll count for lsifstore.BulkMonikerResults. wbnt=%d hbve=%d", 1, len(history))
		} else {
			if diff := cmp.Diff([]int{50, 51, 52, 53, 151, 152, 153}, history[0].Arg2); diff != "" {
				t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
			}

			expectedMonikers := []precise.MonikerDbtb{
				{Kind: "", Scheme: "tsc", Identifier: "tsc npm leftpbd 0.1.0 pbdLeft."},
				{Kind: "", Scheme: "tsc", Identifier: "tsc npm leftpbd 0.2.0 pbd-left."},
			}
			if diff := cmp.Diff(expectedMonikers, history[0].Arg4); diff != "" {
				t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
			}
		}
	})
}

func TestNewGetReferences(t *testing.T) {
	t.Run("locbl", func(t *testing.T) {
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
		mockLsifStore.ExtrbctReferenceLocbtionsFromPositionFunc.PushReturn(locbtions[:1], nil, nil)
		mockLsifStore.ExtrbctReferenceLocbtionsFromPositionFunc.PushReturn(locbtions[1:4], nil, nil)
		mockLsifStore.ExtrbctReferenceLocbtionsFromPositionFunc.PushReturn(locbtions[4:], nil, nil)

		mockCursor := Cursor{}
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
		bdjustedLocbtions, _, err := svc.NewGetReferences(context.Bbckground(), mockRequest, mockRequestStbte, mockCursor)
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
	})

	t.Run("remote", func(t *testing.T) {
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
			{Scheme: "tsc", Identifier: "tsc npm leftpbd 0.1.0 pbdLeft."},
			{Scheme: "tsc", Identifier: "tsc npm leftpbd 0.2.0 pbd_left."},
			{Scheme: "tsc", Identifier: "tsc npm leftpbd 0.3.0 pbd-left."},
		}
		// mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[0]}}, nil)
		// mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[1]}}, nil)
		// mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[2]}}, nil)
		// mockLsifStore.GetMonikersByPositionFunc.PushReturn([][]precise.MonikerDbtb{{monikers[3]}}, nil)

		pbckbgeInformbtion1 := precise.PbckbgeInformbtionDbtb{Mbnbger: "npm", Nbme: "leftpbd", Version: "0.1.0"}
		pbckbgeInformbtion2 := precise.PbckbgeInformbtionDbtb{Mbnbger: "npm", Nbme: "leftpbd", Version: "0.2.0"}
		pbckbgeInformbtion3 := precise.PbckbgeInformbtionDbtb{Mbnbger: "npm", Nbme: "leftpbd", Version: "0.3.0"}
		// mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion1, true, nil)
		// mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion2, true, nil)
		// mockLsifStore.GetPbckbgeInformbtionFunc.PushReturn(pbckbgeInformbtion3, true, nil)

		locbtions := []shbred.Locbtion{
			{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge1},
			{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge2},
			{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge3},
			{DumpID: 51, Pbth: "b.go", Rbnge: testRbnge4},
			{DumpID: 51, Pbth: "c.go", Rbnge: testRbnge5},
		}
		symbolNbmes := []string{
			"tsc npm leftpbd 0.1.0 pbdLeft.",
			"tsc npm leftpbd 0.2.0 pbd_left.",
			"tsc npm leftpbd 0.3.0 pbd-left.",
		}
		mockLsifStore.ExtrbctReferenceLocbtionsFromPositionFunc.PushReturn(locbtions, symbolNbmes, nil)

		monikerLocbtions := []shbred.Locbtion{
			{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge1},
			{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge2},
			{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge3},
			{DumpID: 53, Pbth: "b.go", Rbnge: testRbnge4},
			{DumpID: 53, Pbth: "c.go", Rbnge: testRbnge5},
		}
		mockLsifStore.GetMinimblBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[0:1], 1, nil) // defs
		mockLsifStore.GetMinimblBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[1:2], 1, nil) // refs bbtch 1
		mockLsifStore.GetMinimblBulkMonikerLocbtionsFunc.PushReturn(monikerLocbtions[2:], 3, nil)  // refs bbtch 2

		// uplobds := []dbstore.Dump{
		// 	{ID: 50, Commit: "debdbeef", Root: "sub1/"},
		// 	{ID: 51, Commit: "debdbeef", Root: "sub2/"},
		// 	{ID: 52, Commit: "debdbeef", Root: "sub3/"},
		// 	{ID: 53, Commit: "debdbeef", Root: "sub4/"},
		// }
		// resolver.SetUplobdsDbtbLobder(uplobds)

		mockCursor := Cursor{}
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
		bdjustedLocbtions, _, err := svc.NewGetReferences(context.Bbckground(), mockRequest, mockRequestStbte, mockCursor)
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

		if history := mockLsifStore.GetMinimblBulkMonikerLocbtionsFunc.History(); len(history) != 3 {
			t.Fbtblf("unexpected cbll count for lsifstore.BulkMonikerResults. wbnt=%d hbve=%d", 3, len(history))
		} else {
			if diff := cmp.Diff([]int{50, 51, 52, 53, 151, 152, 153}, history[0].Arg2); diff != "" {
				t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
			}

			expectedMonikers := []precise.MonikerDbtb{
				monikers[0],
				monikers[1],
				monikers[2],
			}
			if diff := cmp.Diff(expectedMonikers, history[0].Arg4); diff != "" {
				t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff([]int{250, 251}, history[1].Arg2); diff != "" {
				t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
			}
			if diff := cmp.Diff(expectedMonikers, history[1].Arg4); diff != "" {
				t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
			}

			if diff := cmp.Diff([]int{252, 253}, history[2].Arg2); diff != "" {
				t.Errorf("unexpected ids (-wbnt +got):\n%s", diff)
			}
			if diff := cmp.Diff(expectedMonikers, history[2].Arg4); diff != "" {
				t.Errorf("unexpected monikers (-wbnt +got):\n%s", diff)
			}
		}
	})
}

func TestNewGetImplementbtions(t *testing.T) {
	t.Run("locbl", func(t *testing.T) {
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
		mockLsifStore.ExtrbctImplementbtionLocbtionsFromPositionFunc.PushReturn(locbtions, nil, nil)

		uplobds := []uplobdsshbred.Dump{
			{ID: 50, Commit: "debdbeef", Root: "sub1/"},
			{ID: 51, Commit: "debdbeef", Root: "sub2/"},
			{ID: 52, Commit: "debdbeef", Root: "sub3/"},
			{ID: 53, Commit: "debdbeef", Root: "sub4/"},
		}
		mockRequestStbte.SetUplobdsDbtbLobder(uplobds)
		mockCursor := Cursor{}
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
		bdjustedLocbtions, _, err := svc.NewGetImplementbtions(context.Bbckground(), mockRequest, mockRequestStbte, mockCursor)
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
	})
}
