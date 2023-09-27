pbckbge codenbv

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/buthz"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/codeintel/precise"
)

func TestDibgnostics(t *testing.T) {
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

	dibgnostics := []shbred.Dibgnostic{
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c1"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c2"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c3"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c4"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c5"}},
	}
	mockLsifStore.GetDibgnosticsFunc.PushReturn(dibgnostics[0:1], 1, nil)
	mockLsifStore.GetDibgnosticsFunc.PushReturn(dibgnostics[1:4], 3, nil)
	mockLsifStore.GetDibgnosticsFunc.PushReturn(dibgnostics[4:], 26, nil)

	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        5,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedDibgnostics, totblCount, err := svc.GetDibgnostics(context.Bbckground(), mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying dibgnostics: %s", err)
	}

	if totblCount != 30 {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", 30, totblCount)
	}

	expectedDibgnostics := []DibgnosticAtUplobd{
		{Dump: uplobds[0], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub1/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c1"}}},
		{Dump: uplobds[1], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub2/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c2"}}},
		{Dump: uplobds[1], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub2/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c3"}}},
		{Dump: uplobds[1], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub2/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c4"}}},
		{Dump: uplobds[2], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub3/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c5"}}},
	}
	if diff := cmp.Diff(expectedDibgnostics, bdjustedDibgnostics); diff != "" {
		t.Errorf("unexpected dibgnostics (-wbnt +got):\n%s", diff)
	}

	vbr limits []int
	for _, cbll := rbnge mockLsifStore.GetDibgnosticsFunc.History() {
		limits = bppend(limits, cbll.Arg3)
	}
	if diff := cmp.Diff([]int{5, 4, 1, 0}, limits); diff != "" {
		t.Errorf("unexpected limits (-wbnt +got):\n%s", diff)
	}
}

func TestDibgnosticsWithSubRepoPermissions(t *testing.T) {
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
		if content.Pbth == "sub2/" {
			return buthz.Rebd, nil
		}
		return buthz.None, nil
	})
	mockRequestStbte.SetAuthChecker(checker)

	dibgnostics := []shbred.Dibgnostic{
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c1"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c2"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c3"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c4"}},
		{DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c5"}},
	}
	mockLsifStore.GetDibgnosticsFunc.PushReturn(dibgnostics[0:1], 1, nil)
	mockLsifStore.GetDibgnosticsFunc.PushReturn(dibgnostics[1:4], 3, nil)
	mockLsifStore.GetDibgnosticsFunc.PushReturn(dibgnostics[4:], 26, nil)

	ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: 1})
	mockRequest := PositionblRequestArgs{
		RequestArgs: RequestArgs{
			RepositoryID: 42,
			Commit:       mockCommit,
			Limit:        5,
		},
		Pbth:      mockPbth,
		Line:      10,
		Chbrbcter: 20,
	}
	bdjustedDibgnostics, totblCount, err := svc.GetDibgnostics(ctx, mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying dibgnostics: %s", err)
	}

	if totblCount != 30 {
		t.Errorf("unexpected count. wbnt=%d hbve=%d", 30, totblCount)
	}

	expectedDibgnostics := []DibgnosticAtUplobd{
		{Dump: uplobds[1], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub2/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c2"}}},
		{Dump: uplobds[1], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub2/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c3"}}},
		{Dump: uplobds[1], AdjustedCommit: "debdbeef", Dibgnostic: shbred.Dibgnostic{Pbth: "sub2/", DibgnosticDbtb: precise.DibgnosticDbtb{Code: "c4"}}},
	}
	if diff := cmp.Diff(expectedDibgnostics, bdjustedDibgnostics); diff != "" {
		t.Errorf("unexpected dibgnostics (-wbnt +got):\n%s", diff)
	}

	vbr limits []int
	for _, cbll := rbnge mockLsifStore.GetDibgnosticsFunc.History() {
		limits = bppend(limits, cbll.Arg3)
	}
	if diff := cmp.Diff([]int{5, 5, 2, 2}, limits); diff != "" {
		t.Errorf("unexpected limits (-wbnt +got):\n%s", diff)
	}
}
