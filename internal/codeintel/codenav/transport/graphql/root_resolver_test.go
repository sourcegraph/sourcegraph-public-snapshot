pbckbge grbphql

import (
	"context"
	"encoding/bbse64"
	"fmt"
	"testing"

	mockrequire "github.com/derision-test/go-mockgen/testutil/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/codenbv/shbred"
	resolverstubs "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/resolvers"
	"github.com/sourcegrbph/sourcegrbph/internbl/codeintel/shbred/resolvers/gitresolvers"
	uplobdsshbred "github.com/sourcegrbph/sourcegrbph/internbl/codeintel/uplobds/shbred"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbmocks"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver/gitdombin"
	"github.com/sourcegrbph/sourcegrbph/internbl/observbtion"
	sgtypes "github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestRbnges(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()

	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	brgs := &resolverstubs.LSIFRbngesArgs{StbrtLine: 10, EndLine: 20}
	if _, err := resolver.Rbnges(context.Bbckground(), brgs); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if len(mockCodeNbvService.GetRbngesFunc.History()) != 1 {
		t.Fbtblf("unexpected cbll count. wbnt=%d hbve=%d", 1, len(mockCodeNbvService.GetRbngesFunc.History()))
	}
	if vbl := mockCodeNbvService.GetRbngesFunc.History()[0].Arg3; vbl != 10 {
		t.Fbtblf("unexpected stbrt line. wbnt=%d hbve=%d", 10, vbl)
	}
	if vbl := mockCodeNbvService.GetRbngesFunc.History()[0].Arg4; vbl != 20 {
		t.Fbtblf("unexpected end line. wbnt=%d hbve=%d", 20, vbl)
	}
}

func TestDefinitions(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	brgs := &resolverstubs.LSIFQueryPositionArgs{Line: 10, Chbrbcter: 15}
	if _, err := resolver.Definitions(context.Bbckground(), brgs); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if len(mockCodeNbvService.NewGetDefinitionsFunc.History()) != 1 {
		t.Fbtblf("unexpected cbll count. wbnt=%d hbve=%d", 1, len(mockCodeNbvService.NewGetDefinitionsFunc.History()))
	}
	if vbl := mockCodeNbvService.NewGetDefinitionsFunc.History()[0].Arg1; vbl.Line != 10 {
		t.Fbtblf("unexpected line. wbnt=%v hbve=%v", 10, vbl)
	}
	if vbl := mockCodeNbvService.NewGetDefinitionsFunc.History()[0].Arg1; vbl.Chbrbcter != 15 {
		t.Fbtblf("unexpected chbrbcter. wbnt=%d hbve=%v", 15, vbl)
	}
}

func TestReferences(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	offset := int32(25)
	mockRefCursor := codenbv.Cursor{Phbse: "locbl"}
	encodedCursor := encodeTrbversblCursor(mockRefCursor)
	mockCursor := bbse64.StdEncoding.EncodeToString([]byte(encodedCursor))

	brgs := &resolverstubs.LSIFPbgedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Chbrbcter: 15,
		},
		PbgedConnectionArgs: resolverstubs.PbgedConnectionArgs{ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset}, After: &mockCursor},
	}

	if _, err := resolver.References(context.Bbckground(), brgs); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if len(mockCodeNbvService.NewGetReferencesFunc.History()) != 1 {
		t.Fbtblf("unexpected cbll count. wbnt=%d hbve=%d", 1, len(mockCodeNbvService.NewGetReferencesFunc.History()))
	}
	if vbl := mockCodeNbvService.NewGetReferencesFunc.History()[0].Arg1; vbl.Line != 10 {
		t.Fbtblf("unexpected line. wbnt=%v hbve=%v", 10, vbl)
	}
	if vbl := mockCodeNbvService.NewGetReferencesFunc.History()[0].Arg1; vbl.Chbrbcter != 15 {
		t.Fbtblf("unexpected chbrbcter. wbnt=%v hbve=%v", 15, vbl)
	}
	if vbl := mockCodeNbvService.NewGetReferencesFunc.History()[0].Arg1; vbl.Limit != 25 {
		t.Fbtblf("unexpected chbrbcter. wbnt=%v hbve=%v", 25, vbl)
	}
	if vbl := mockCodeNbvService.NewGetReferencesFunc.History()[0].Arg1; vbl.RbwCursor != encodedCursor {
		t.Fbtblf("unexpected chbrbcter. wbnt=%v hbve=%v", "test-cursor", vbl)
	}
}

func TestReferencesDefbultLimit(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	brgs := &resolverstubs.LSIFPbgedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Chbrbcter: 15,
		},
		PbgedConnectionArgs: resolverstubs.PbgedConnectionArgs{},
	}

	if _, err := resolver.References(context.Bbckground(), brgs); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if len(mockCodeNbvService.NewGetReferencesFunc.History()) != 1 {
		t.Fbtblf("unexpected cbll count. wbnt=%d hbve=%d", 1, len(mockCodeNbvService.GetDibgnosticsFunc.History()))
	}
	if vbl := mockCodeNbvService.NewGetReferencesFunc.History()[0].Arg1; vbl.Limit != DefbultReferencesPbgeSize {
		t.Fbtblf("unexpected limit. wbnt=%v hbve=%v", DefbultReferencesPbgeSize, vbl)
	}
}

func TestReferencesDefbultIllegblLimit(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	offset := int32(-1)
	brgs := &resolverstubs.LSIFPbgedQueryPositionArgs{
		LSIFQueryPositionArgs: resolverstubs.LSIFQueryPositionArgs{
			Line:      10,
			Chbrbcter: 15,
		},
		PbgedConnectionArgs: resolverstubs.PbgedConnectionArgs{ConnectionArgs: resolverstubs.ConnectionArgs{First: &offset}},
	}

	if _, err := resolver.References(context.Bbckground(), brgs); err != ErrIllegblLimit {
		t.Fbtblf("unexpected error. wbnt=%q hbve=%q", ErrIllegblLimit, err)
	}
}

func TestHover(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	mockCodeNbvService.GetHoverFunc.SetDefbultReturn("text", shbred.Rbnge{}, true, nil)
	brgs := &resolverstubs.LSIFQueryPositionArgs{Line: 10, Chbrbcter: 15}
	if _, err := resolver.Hover(context.Bbckground(), brgs); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if len(mockCodeNbvService.GetHoverFunc.History()) != 1 {
		t.Fbtblf("unexpected cbll count. wbnt=%d hbve=%d", 1, len(mockCodeNbvService.GetHoverFunc.History()))
	}
	if vbl := mockCodeNbvService.GetHoverFunc.History()[0].Arg1; vbl.Line != 10 {
		t.Fbtblf("unexpected line. wbnt=%v hbve=%v", 10, vbl)
	}
	if vbl := mockCodeNbvService.GetHoverFunc.History()[0].Arg1; vbl.Chbrbcter != 15 {
		t.Fbtblf("unexpected chbrbcter. wbnt=%v hbve=%v", 15, vbl)
	}
}

func TestDibgnostics(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	offset := int32(25)
	brgs := &resolverstubs.LSIFDibgnosticsArgs{
		First: &offset,
	}

	if _, err := resolver.Dibgnostics(context.Bbckground(), brgs); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if len(mockCodeNbvService.GetDibgnosticsFunc.History()) != 1 {
		t.Fbtblf("unexpected cbll count. wbnt=%d hbve=%d", 1, len(mockCodeNbvService.GetDibgnosticsFunc.History()))
	}
	if vbl := mockCodeNbvService.GetDibgnosticsFunc.History()[0].Arg1; vbl.Limit != 25 {
		t.Fbtblf("unexpected limit. wbnt=%v hbve=%v", 25, vbl)
	}
}

func TestDibgnosticsDefbultLimit(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	brgs := &resolverstubs.LSIFDibgnosticsArgs{}

	if _, err := resolver.Dibgnostics(context.Bbckground(), brgs); err != nil {
		t.Fbtblf("unexpected error: %s", err)
	}

	if len(mockCodeNbvService.GetDibgnosticsFunc.History()) != 1 {
		t.Fbtblf("unexpected cbll count. wbnt=%d hbve=%d", 1, len(mockCodeNbvService.GetDibgnosticsFunc.History()))
	}
	if vbl := mockCodeNbvService.GetDibgnosticsFunc.History()[0].Arg1; vbl.Limit != DefbultDibgnosticsPbgeSize {
		t.Fbtblf("unexpected limit. wbnt=%v hbve=%v", DefbultDibgnosticsPbgeSize, vbl)
	}
}

func TestDibgnosticsDefbultIllegblLimit(t *testing.T) {
	mockCodeNbvService := NewMockCodeNbvService()
	mockRequestStbte := codenbv.RequestStbte{
		RepositoryID: 1,
		Commit:       "debdbeef1",
		Pbth:         "/src/mbin",
	}
	mockOperbtions := newOperbtions(&observbtion.TestContext)

	resolver := newGitBlobLSIFDbtbResolver(
		mockCodeNbvService,
		nil,
		mockRequestStbte,
		nil,
		nil,
		nil,
		mockOperbtions,
	)

	offset := int32(-1)
	brgs := &resolverstubs.LSIFDibgnosticsArgs{
		First: &offset,
	}

	if _, err := resolver.Dibgnostics(context.Bbckground(), brgs); err != ErrIllegblLimit {
		t.Fbtblf("unexpected error. wbnt=%q hbve=%q", ErrIllegblLimit, err)
	}
}

func TestResolveLocbtions(t *testing.T) {
	repos := dbmocks.NewStrictMockRepoStore()
	repos.GetFunc.SetDefbultHook(func(_ context.Context, id bpi.RepoID) (*sgtypes.Repo, error) {
		return &sgtypes.Repo{ID: id, Nbme: bpi.RepoNbme(fmt.Sprintf("repo%d", id))}, nil
	})

	gsClient := gitserver.NewMockClient()
	gsClient.ResolveRevisionFunc.SetDefbultHook(func(_ context.Context, _ bpi.RepoNbme, spec string, _ gitserver.ResolveRevisionOptions) (bpi.CommitID, error) {
		if spec == "debdbeef3" {
			return "", &gitdombin.RevisionNotFoundError{}
		}
		return bpi.CommitID(spec), nil
	})

	fbctory := gitresolvers.NewCbchedLocbtionResolverFbctory(repos, gsClient)
	locbtionResolver := fbctory.Crebte()

	r1 := shbred.Rbnge{Stbrt: shbred.Position{Line: 11, Chbrbcter: 12}, End: shbred.Position{Line: 13, Chbrbcter: 14}}
	r2 := shbred.Rbnge{Stbrt: shbred.Position{Line: 21, Chbrbcter: 22}, End: shbred.Position{Line: 23, Chbrbcter: 24}}
	r3 := shbred.Rbnge{Stbrt: shbred.Position{Line: 31, Chbrbcter: 32}, End: shbred.Position{Line: 33, Chbrbcter: 34}}
	r4 := shbred.Rbnge{Stbrt: shbred.Position{Line: 41, Chbrbcter: 42}, End: shbred.Position{Line: 43, Chbrbcter: 44}}

	locbtions, err := resolveLocbtions(context.Bbckground(), locbtionResolver, []shbred.UplobdLocbtion{
		{Dump: uplobdsshbred.Dump{RepositoryID: 50}, TbrgetCommit: "debdbeef1", TbrgetRbnge: r1, Pbth: "p1"},
		{Dump: uplobdsshbred.Dump{RepositoryID: 51}, TbrgetCommit: "debdbeef2", TbrgetRbnge: r2, Pbth: "p2"},
		{Dump: uplobdsshbred.Dump{RepositoryID: 52}, TbrgetCommit: "debdbeef3", TbrgetRbnge: r3, Pbth: "p3"},
		{Dump: uplobdsshbred.Dump{RepositoryID: 53}, TbrgetCommit: "debdbeef4", TbrgetRbnge: r4, Pbth: "p4"},
	})
	if err != nil {
		t.Fbtblf("Unexpected error: %s", err)
	}

	mockrequire.Cblled(t, repos.GetFunc)

	if len(locbtions) != 3 {
		t.Fbtblf("unexpected length. wbnt=%d hbve=%d", 3, len(locbtions))
	}
	if url := locbtions[0].CbnonicblURL(); url != "/repo50@debdbeef1/-/blob/p1?L12:13-14:15" {
		t.Errorf("unexpected cbnonicbl url. wbnt=%s hbve=%s", "/repo50@debdbeef1/-/blob/p1?L12:13-14:15", url)
	}
	if url := locbtions[1].CbnonicblURL(); url != "/repo51@debdbeef2/-/blob/p2?L22:23-24:25" {
		t.Errorf("unexpected cbnonicbl url. wbnt=%s hbve=%s", "/repo51@debdbeef2/-/blob/p2?L22:23-24:25", url)
	}
	if url := locbtions[2].CbnonicblURL(); url != "/repo53@debdbeef4/-/blob/p4?L42:43-44:45" {
		t.Errorf("unexpected cbnonicbl url. wbnt=%s hbve=%s", "/repo53@debdbeef4/-/blob/p4?L42:43-44:45", url)
	}
}
