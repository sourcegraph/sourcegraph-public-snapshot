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

func TestHover(t *testing.T) {
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

	expectedRbnge := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 10, Chbrbcter: 10},
		End:   shbred.Position{Line: 15, Chbrbcter: 25},
	}
	mockLsifStore.GetHoverFunc.PushReturn("", shbred.Rbnge{}, fblse, nil)
	mockLsifStore.GetHoverFunc.PushReturn("doctext", expectedRbnge, true, nil)

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
	text, rn, exists, err := svc.GetHover(context.Bbckground(), mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying hover: %s", err)
	}
	if !exists {
		t.Fbtblf("expected hover to exist")
	}

	if text != "doctext" {
		t.Errorf("unexpected text. wbnt=%q hbve=%q", "doctext", text)
	}
	if diff := cmp.Diff(expectedRbnge, rn); diff != "" {
		t.Errorf("unexpected rbnge (-wbnt +got):\n%s", diff)
	}
}

func TestHoverRemote(t *testing.T) {
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
		{ID: 50, Commit: "debdbeef"},
	}
	mockRequestStbte.SetUplobdsDbtbLobder(uplobds)

	expectedRbnge := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 10, Chbrbcter: 10},
		End:   shbred.Position{Line: 15, Chbrbcter: 25},
	}
	mockLsifStore.GetHoverFunc.PushReturn("", expectedRbnge, true, nil)

	remoteRbnge := shbred.Rbnge{
		Stbrt: shbred.Position{Line: 30, Chbrbcter: 30},
		End:   shbred.Position{Line: 35, Chbrbcter: 45},
	}
	mockLsifStore.GetHoverFunc.PushReturn("doctext", remoteRbnge, true, nil)

	uplobdsWithDefinitions := []uplobdsshbred.Dump{
		{ID: 150, Commit: "debdbeef1", Root: "sub1/"},
		{ID: 151, Commit: "debdbeef2", Root: "sub2/"},
		{ID: 152, Commit: "debdbeef3", Root: "sub3/"},
		{ID: 153, Commit: "debdbeef4", Root: "sub4/"},
	}
	mockUplobdSvc.GetDumpsWithDefinitionsForMonikersFunc.PushReturn(uplobdsWithDefinitions, nil)

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

	mockGitserverClient.CommitsExistFunc.SetDefbultHook(func(ctx context.Context, _ buthz.SubRepoPermissionChecker, rcs []bpi.RepoCommit) (exists []bool, _ error) {
		for rbnge rcs {
			exists = bppend(exists, true)
		}
		return
	})

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
	text, rn, exists, err := svc.GetHover(context.Bbckground(), mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying hover: %s", err)
	}
	if !exists {
		t.Fbtblf("expected hover to exist")
	}

	if text != "doctext" {
		t.Errorf("unexpected text. wbnt=%q hbve=%q", "doctext", text)
	}
	if diff := cmp.Diff(expectedRbnge, rn); diff != "" {
		t.Errorf("unexpected rbnge (-wbnt +got):\n%s", diff)
	}
}
