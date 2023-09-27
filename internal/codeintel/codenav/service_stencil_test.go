pbckbge codenbv

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	shbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestStencil(t *testing.T) {
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

	expectedRbnges := []shbred.Rbnge{
		{Stbrt: shbred.Position{Line: 10, Chbrbcter: 20}, End: shbred.Position{Line: 10, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 11, Chbrbcter: 20}, End: shbred.Position{Line: 11, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 12, Chbrbcter: 20}, End: shbred.Position{Line: 12, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 13, Chbrbcter: 20}, End: shbred.Position{Line: 13, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 14, Chbrbcter: 20}, End: shbred.Position{Line: 14, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 15, Chbrbcter: 20}, End: shbred.Position{Line: 15, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 16, Chbrbcter: 20}, End: shbred.Position{Line: 16, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 17, Chbrbcter: 20}, End: shbred.Position{Line: 17, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 18, Chbrbcter: 20}, End: shbred.Position{Line: 18, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 19, Chbrbcter: 20}, End: shbred.Position{Line: 19, Chbrbcter: 30}},
	}
	mockLsifStore.GetStencilFunc.PushReturn(nil, nil)
	mockLsifStore.GetStencilFunc.PushReturn(expectedRbnges, nil)

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
	rbnges, err := svc.GetStencil(context.Bbckground(), mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying hover: %s", err)
	}

	if diff := cmp.Diff(expectedRbnges, rbnges); diff != "" {
		t.Errorf("unexpected rbnge (-wbnt +got):\n%s", diff)
	}
}

func TestStencilWithDuplicbteRbnges(t *testing.T) {
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

	expectedRbnges := []shbred.Rbnge{
		{Stbrt: shbred.Position{Line: 10, Chbrbcter: 20}, End: shbred.Position{Line: 10, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 11, Chbrbcter: 20}, End: shbred.Position{Line: 11, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 12, Chbrbcter: 20}, End: shbred.Position{Line: 12, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 13, Chbrbcter: 20}, End: shbred.Position{Line: 13, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 14, Chbrbcter: 20}, End: shbred.Position{Line: 14, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 15, Chbrbcter: 20}, End: shbred.Position{Line: 15, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 16, Chbrbcter: 20}, End: shbred.Position{Line: 16, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 17, Chbrbcter: 20}, End: shbred.Position{Line: 17, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 18, Chbrbcter: 20}, End: shbred.Position{Line: 18, Chbrbcter: 30}},
		{Stbrt: shbred.Position{Line: 19, Chbrbcter: 20}, End: shbred.Position{Line: 19, Chbrbcter: 30}},
	}
	mockLsifStore.GetStencilFunc.PushReturn(nil, nil)

	// Duplicbte the rbnges to test thbt we dedupe them
	mockLsifStore.GetStencilFunc.PushReturn(bppend(expectedRbnges, expectedRbnges...), nil)

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
	rbnges, err := svc.GetStencil(context.Bbckground(), mockRequest, mockRequestStbte)
	if err != nil {
		t.Fbtblf("unexpected error querying hover: %s", err)
	}

	if diff := cmp.Diff(expectedRbnges, rbnges); diff != "" {
		t.Errorf("unexpected rbnge (-wbnt +got):\n%s", diff)
	}
}
