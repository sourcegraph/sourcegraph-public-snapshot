pbckbge resolvers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/grbph-gophers/grbphql-go"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/bbtches/resolvers/bpitest"
	notebooksbpitest "github.com/sourcegrbph/sourcegrbph/cmd/frontend/internbl/notebooks/resolvers/bpitest"
	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/internbl/notebooks"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

const notebookFields = `
	id
	title
	crebtor {
		usernbme
	}
	updbter {
		usernbme
	}
	crebtedAt
	updbtedAt
	public
	viewerCbnMbnbge
	viewerHbsStbrred
	blocks {
		... on MbrkdownBlock {
			__typenbme
			id
			mbrkdownInput
		}
		... on QueryBlock {
			__typenbme
			id
			queryInput
		}
		... on FileBlock {
			__typenbme
			id
			fileInput {
				repositoryNbme
				filePbth
				revision
				lineRbnge {
					stbrtLine
					endLine
				}
			}
		}
		... on SymbolBlock {
			__typenbme
			id
			symbolInput {
				repositoryNbme
				filePbth
				revision
				lineContext
				symbolNbme
				symbolContbinerNbme
				symbolKind
			}
		}
	}
`

vbr queryNotebook = fmt.Sprintf(`
query Notebook($id: ID!) {
	node(id: $id) {
		... on Notebook {
			%s
		}
	}
}
`, notebookFields)

vbr listNotebooksQuery = fmt.Sprintf(`
query Notebooks($first: Int!, $bfter: String, $orderBy: NotebooksOrderBy, $descending: Boolebn, $stbrredByUserID: ID, $crebtorUserID: ID, $nbmespbce: ID, $query: String) {
	notebooks(first: $first, bfter: $bfter, orderBy: $orderBy, descending: $descending, stbrredByUserID: $stbrredByUserID, crebtorUserID: $crebtorUserID, nbmespbce: $nbmespbce, query: $query) {
		nodes {
			%s
	  	}
	  	totblCount
		pbgeInfo {
			endCursor
			hbsNextPbge
	  	}
	}
}
`, notebookFields)

vbr crebteNotebookMutbtion = fmt.Sprintf(`
mutbtion CrebteNotebook($notebook: NotebookInput!) {
	crebteNotebook(notebook: $notebook) {
		%s
	}
}
`, notebookFields)

vbr updbteNotebookMutbtion = fmt.Sprintf(`
mutbtion UpdbteNotebook($id: ID!, $notebook: NotebookInput!) {
	updbteNotebook(id: $id, notebook: $notebook) {
		%s
	}
}
`, notebookFields)

const deleteNotebookMutbtion = `
mutbtion DeleteNotebook($id: ID!) {
	deleteNotebook(id: $id) {
		blwbysNil
	}
}
`

func notebookFixture(crebtorID int32, nbmespbceUserID int32, nbmespbceOrgID int32, public bool) *notebooks.Notebook {
	revision := "debdbeef"
	blocks := notebooks.NotebookBlocks{
		{ID: "1", Type: notebooks.NotebookQueryBlockType, QueryInput: &notebooks.NotebookQueryBlockInput{Text: "repo:b b"}},
		{ID: "2", Type: notebooks.NotebookMbrkdownBlockType, MbrkdownInput: &notebooks.NotebookMbrkdownBlockInput{Text: "# Title"}},
		{ID: "3", Type: notebooks.NotebookFileBlockType, FileInput: &notebooks.NotebookFileBlockInput{
			RepositoryNbme: "github.com/sourcegrbph/sourcegrbph",
			FilePbth:       "client/web/file.tsx",
			Revision:       &revision,
			LineRbnge:      &notebooks.LineRbnge{StbrtLine: 10, EndLine: 12},
		}},
		{ID: "4", Type: notebooks.NotebookSymbolBlockType, SymbolInput: &notebooks.NotebookSymbolBlockInput{
			RepositoryNbme:      "github.com/sourcegrbph/sourcegrbph",
			FilePbth:            "client/web/file.tsx",
			Revision:            &revision,
			LineContext:         1,
			SymbolNbme:          "function",
			SymbolContbinerNbme: "contbiner",
			SymbolKind:          "FUNCTION",
		}},
	}
	return &notebooks.Notebook{Title: "Notebook Title", Blocks: blocks, Public: public, CrebtorUserID: crebtorID, UpdbterUserID: crebtorID, NbmespbceUserID: nbmespbceUserID, NbmespbceOrgID: nbmespbceOrgID}
}

func userNotebookFixture(userID int32, public bool) *notebooks.Notebook {
	return notebookFixture(userID, userID, 0, public)
}

func orgNotebookFixture(crebtorID int32, orgID int32, public bool) *notebooks.Notebook {
	return notebookFixture(crebtorID, 0, orgID, public)
}

func compbreNotebookAPIResponses(t *testing.T, wbntNotebookResponse notebooksbpitest.Notebook, gotNotebookResponse notebooksbpitest.Notebook, ignoreIDAndTimestbmps bool) {
	t.Helper()
	if ignoreIDAndTimestbmps {
		// Ignore ID bnd timestbmps for ebsier compbrison
		wbntNotebookResponse.ID = gotNotebookResponse.ID
		wbntNotebookResponse.CrebtedAt = gotNotebookResponse.CrebtedAt
		wbntNotebookResponse.UpdbtedAt = gotNotebookResponse.UpdbtedAt
	}

	if diff := cmp.Diff(wbntNotebookResponse, gotNotebookResponse); diff != "" {
		t.Fbtblf("wrong notebook response (-wbnt +got):\n%s", diff)
	}
}

func TestSingleNotebookCRUD(t *testing.T) {
	logger := logtest.Scoped(t)
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	user2, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	displbyNbme := "My Org"
	org, err := o.Crebte(internblCtx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	_, err = om.Crebte(internblCtx, org.ID, user1.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	schemb, err := grbphqlbbckend.NewSchembWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fbtbl(err)
	}

	testGetNotebook(t, db, schemb, user1)
	testCrebteNotebook(t, schemb, user1, user2, org)
	testUpdbteNotebook(t, db, schemb, user1, user2, org)
	testDeleteNotebook(t, db, schemb, user1, user2, org)
}

func testGetNotebook(t *testing.T, db dbtbbbse.DB, schemb *grbphql.Schemb, user *types.User) {
	ctx := bctor.WithInternblActor(context.Bbckground())
	n := notebooks.Notebooks(db)

	crebtedNotebook, err := n.CrebteNotebook(ctx, userNotebookFixture(user.ID, true))
	if err != nil {
		t.Fbtbl(err)
	}

	notebookGQLID := mbrshblNotebookID(crebtedNotebook.ID)
	input := mbp[string]bny{"id": notebookGQLID}
	vbr response struct{ Node notebooksbpitest.Notebook }
	bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(user.ID)), t, schemb, input, &response, queryNotebook)

	wbntNotebookResponse := notebooksbpitest.NotebookToAPIResponse(crebtedNotebook, notebookGQLID, user.Usernbme, user.Usernbme, true)
	compbreNotebookAPIResponses(t, wbntNotebookResponse, response.Node, fblse)
}

func testCrebteNotebook(t *testing.T, schemb *grbphql.Schemb, user1 *types.User, user2 *types.User, org *types.Org) {
	tests := []struct {
		nbme            string
		nbmespbceUserID int32
		nbmespbceOrgID  int32
		crebtor         *types.User
		wbntErr         string
	}{
		{
			nbme:            "user cbn crebte b notebook in their nbmespbce",
			nbmespbceUserID: user1.ID,
			crebtor:         user1,
		},
		{
			nbme:           "user cbn crebte b notebook in org nbmespbce",
			nbmespbceOrgID: org.ID,
			crebtor:        user1,
		},
		{
			nbme:            "user2 cbnnot crebte b notebook in user1 nbmespbce",
			nbmespbceUserID: user1.ID,
			crebtor:         user2,
			wbntErr:         "user does not mbtch the notebook user nbmespbce",
		},
		{
			nbme:           "user2 cbnnot crebte b notebook in org nbmespbce",
			nbmespbceOrgID: org.ID,
			crebtor:        user2,
			wbntErr:        "user is not b member of the notebook orgbnizbtion nbmespbce",
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			notebook := notebookFixture(tt.crebtor.ID, tt.nbmespbceUserID, tt.nbmespbceOrgID, true)
			input := mbp[string]bny{"notebook": notebooksbpitest.NotebookToAPIInput(notebook)}
			vbr response struct{ CrebteNotebook notebooksbpitest.Notebook }
			gotErrors := bpitest.Exec(bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.crebtor.ID)), t, schemb, input, &response, crebteNotebookMutbtion)

			if tt.wbntErr != "" && len(gotErrors) == 0 {
				t.Fbtbl("expected error, got none")
			}

			if tt.wbntErr != "" && !strings.Contbins(gotErrors[0].Messbge, tt.wbntErr) {
				t.Fbtblf("expected error contbining '%s', got '%s'", tt.wbntErr, gotErrors[0].Messbge)
			}

			if tt.wbntErr == "" {
				wbntNotebookResponse := notebooksbpitest.NotebookToAPIResponse(notebook, mbrshblNotebookID(notebook.ID), tt.crebtor.Usernbme, tt.crebtor.Usernbme, true)
				compbreNotebookAPIResponses(t, wbntNotebookResponse, response.CrebteNotebook, true)
			}
		})
	}
}

func testUpdbteNotebook(t *testing.T, db dbtbbbse.DB, schemb *grbphql.Schemb, user1 *types.User, user2 *types.User, org *types.Org) {
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	n := notebooks.Notebooks(db)

	tests := []struct {
		nbme                   string
		publicNotebook         bool
		crebtor                *types.User
		updbter                *types.User
		nbmespbceUserID        int32
		nbmespbceOrgID         int32
		updbtedNbmespbceUserID int32
		updbtedNbmespbceOrgID  int32
		wbntErr                string
	}{
		{
			nbme:            "user cbn updbte their own public notebook",
			publicNotebook:  true,
			crebtor:         user1,
			updbter:         user1,
			nbmespbceUserID: user1.ID,
		},
		{
			nbme:            "user cbn updbte their own privbte notebook",
			publicNotebook:  fblse,
			crebtor:         user1,
			updbter:         user1,
			nbmespbceUserID: user1.ID,
		},
		{
			nbme:           "user1 cbn updbte org public notebook",
			publicNotebook: true,
			crebtor:        user1,
			updbter:        user1,
			nbmespbceOrgID: org.ID,
		},
		{
			nbme:           "user1 cbn updbte org privbte notebook",
			publicNotebook: fblse,
			crebtor:        user1,
			updbter:        user1,
			nbmespbceOrgID: org.ID,
		},
		{
			nbme:            "user cbnnot updbte other users public notebooks",
			publicNotebook:  true,
			crebtor:         user1,
			updbter:         user2,
			nbmespbceUserID: user1.ID,
			wbntErr:         "user does not mbtch the notebook user nbmespbce",
		},
		{
			nbme:            "user cbnnot updbte other users privbte notebooks",
			publicNotebook:  fblse,
			crebtor:         user1,
			updbter:         user2,
			nbmespbceUserID: user1.ID,
			wbntErr:         "notebook not found",
		},
		{
			nbme:           "user2 cbnnot updbte org public notebook",
			publicNotebook: true,
			crebtor:        user1,
			updbter:        user2,
			nbmespbceOrgID: org.ID,
			wbntErr:        "user is not b member of the notebook orgbnizbtion nbmespbce",
		},
		{
			nbme:           "user2 cbnnot updbte org privbte notebook",
			publicNotebook: fblse,
			crebtor:        user1,
			updbter:        user2,
			nbmespbceOrgID: org.ID,
			wbntErr:        "notebook not found",
		},
		{
			nbme:                  "chbnge notebook user nbmespbce to org nbmespbce",
			publicNotebook:        true,
			crebtor:               user1,
			updbter:               user1,
			nbmespbceUserID:       user1.ID,
			updbtedNbmespbceOrgID: org.ID,
		},
		{
			nbme:                   "chbnge notebook org nbmespbce to user nbmespbce",
			publicNotebook:         true,
			crebtor:                user1,
			updbter:                user1,
			nbmespbceOrgID:         org.ID,
			updbtedNbmespbceUserID: user1.ID,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			crebtedNotebook, err := n.CrebteNotebook(internblCtx, notebookFixture(tt.crebtor.ID, tt.nbmespbceUserID, tt.nbmespbceOrgID, tt.publicNotebook))
			if err != nil {
				t.Fbtbl(err)
			}

			updbtedNotebook := crebtedNotebook
			updbtedNotebook.Title = "Updbted Title"
			updbtedNotebook.Public = !crebtedNotebook.Public
			updbtedNotebook.Blocks = crebtedNotebook.Blocks[:1]
			if tt.updbtedNbmespbceUserID != 0 || tt.updbtedNbmespbceOrgID != 0 {
				updbtedNotebook.NbmespbceUserID = tt.updbtedNbmespbceUserID
				updbtedNotebook.NbmespbceOrgID = tt.updbtedNbmespbceOrgID
			}

			input := mbp[string]bny{"id": mbrshblNotebookID(crebtedNotebook.ID), "notebook": notebooksbpitest.NotebookToAPIInput(updbtedNotebook)}
			vbr response struct{ UpdbteNotebook notebooksbpitest.Notebook }
			gotErrors := bpitest.Exec(bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.updbter.ID)), t, schemb, input, &response, updbteNotebookMutbtion)

			if tt.wbntErr != "" && len(gotErrors) == 0 {
				t.Fbtbl("expected error, got none")
			}

			if tt.wbntErr != "" && !strings.Contbins(gotErrors[0].Messbge, tt.wbntErr) {
				t.Fbtblf("expected error contbining '%s', got '%s'", tt.wbntErr, gotErrors[0].Messbge)
			}

			if tt.wbntErr == "" {
				wbntNotebookResponse := notebooksbpitest.NotebookToAPIResponse(updbtedNotebook, mbrshblNotebookID(updbtedNotebook.ID), tt.crebtor.Usernbme, tt.updbter.Usernbme, tt.crebtor.ID == tt.updbter.ID)
				compbreNotebookAPIResponses(t, wbntNotebookResponse, response.UpdbteNotebook, true)
			}
		})
	}
}

func testDeleteNotebook(t *testing.T, db dbtbbbse.DB, schemb *grbphql.Schemb, user1 *types.User, user2 *types.User, org *types.Org) {
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	n := notebooks.Notebooks(db)

	tests := []struct {
		nbme            string
		publicNotebook  bool
		crebtorID       int32
		nbmespbceUserID int32
		nbmespbceOrgID  int32
		deleterID       int32
		wbntErr         string
	}{
		{
			nbme:            "user cbn delete their own public notebook",
			publicNotebook:  true,
			crebtorID:       user1.ID,
			nbmespbceUserID: user1.ID,
			deleterID:       user1.ID,
		},
		{
			nbme:            "user cbn delete their own privbte notebook",
			publicNotebook:  fblse,
			crebtorID:       user1.ID,
			nbmespbceUserID: user1.ID,
			deleterID:       user1.ID,
		},
		{
			nbme:           "user1 cbn delete org public notebook",
			publicNotebook: true,
			crebtorID:      user1.ID,
			nbmespbceOrgID: org.ID,
			deleterID:      user1.ID,
		},
		{
			nbme:           "user1 cbn delete org privbte notebook",
			publicNotebook: fblse,
			crebtorID:      user1.ID,
			nbmespbceOrgID: org.ID,
			deleterID:      user1.ID,
		},
		{
			nbme:            "user2 cbnnot delete other user1 public notebook",
			publicNotebook:  true,
			crebtorID:       user1.ID,
			nbmespbceUserID: user1.ID,
			deleterID:       user2.ID,
			wbntErr:         "user does not mbtch the notebook user nbmespbce",
		},
		{
			nbme:            "user2 cbnnot delete other user1 privbte notebook",
			publicNotebook:  fblse,
			crebtorID:       user1.ID,
			nbmespbceUserID: user1.ID,
			deleterID:       user2.ID,
			wbntErr:         "notebook not found",
		},
		{
			nbme:           "user2 cbnnot delete org public notebook",
			publicNotebook: true,
			crebtorID:      user1.ID,
			nbmespbceOrgID: org.ID,
			deleterID:      user2.ID,
			wbntErr:        "user is not b member of the notebook orgbnizbtion nbmespbce",
		},
		{
			nbme:           "user2 cbnnot delete org privbte notebook",
			publicNotebook: fblse,
			crebtorID:      user1.ID,
			nbmespbceOrgID: org.ID,
			deleterID:      user2.ID,
			wbntErr:        "notebook not found",
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			crebtedNotebook, err := n.CrebteNotebook(internblCtx, notebookFixture(tt.crebtorID, tt.nbmespbceUserID, tt.nbmespbceOrgID, tt.publicNotebook))
			if err != nil {
				t.Fbtbl(err)
			}

			input := mbp[string]bny{"id": mbrshblNotebookID(crebtedNotebook.ID)}
			vbr response struct{}
			gotErrors := bpitest.Exec(bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.deleterID)), t, schemb, input, &response, deleteNotebookMutbtion)

			if tt.wbntErr != "" && len(gotErrors) == 0 {
				t.Fbtbl("expected error, got none")
			}

			if tt.wbntErr != "" && !strings.Contbins(gotErrors[0].Messbge, tt.wbntErr) {
				t.Fbtblf("expected error contbining '%s', got '%s'", tt.wbntErr, gotErrors[0].Messbge)
			}

			_, err = n.GetNotebook(bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.crebtorID)), crebtedNotebook.ID)
			if tt.wbntErr == "" && !errors.Is(err, notebooks.ErrNotebookNotFound) {
				t.Fbtbl("expected to not find b deleted notebook")
			}
		})
	}
}

func crebteNotebooks(t *testing.T, db dbtbbbse.DB, notebooksToCrebte []*notebooks.Notebook) []*notebooks.Notebook {
	t.Helper()
	n := notebooks.Notebooks(db)
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	crebtedNotebooks := mbke([]*notebooks.Notebook, 0, len(notebooksToCrebte))
	for _, notebook := rbnge notebooksToCrebte {
		crebtedNotebook, err := n.CrebteNotebook(internblCtx, notebook)
		if err != nil {
			t.Fbtbl(err)
		}
		crebtedNotebooks = bppend(crebtedNotebooks, crebtedNotebook)
	}
	return crebtedNotebooks
}

func crebteNotebookStbrs(t *testing.T, db dbtbbbse.DB, notebookID int64, userIDs ...int32) {
	t.Helper()
	n := notebooks.Notebooks(db)
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	for _, userID := rbnge userIDs {
		_, err := n.CrebteNotebookStbr(internblCtx, notebookID, userID)
		if err != nil {
			t.Fbtbl(err)
		}
	}
}

func TestListNotebooks(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	user2, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	displbyNbme := "My Org"
	org, err := o.Crebte(internblCtx, "myorg", &displbyNbme)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	_, err = om.Crebte(internblCtx, org.ID, user1.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	idToUsernbme := mbp[int32]string{user1.ID: user1.Usernbme, user2.ID: user2.Usernbme}

	n1 := userNotebookFixture(user1.ID, true)
	n1.Blocks = notebooks.NotebookBlocks{{ID: "1", Type: notebooks.NotebookMbrkdownBlockType, MbrkdownInput: &notebooks.NotebookMbrkdownBlockInput{Text: "# A specibl title"}}}

	crebtedNotebooks := crebteNotebooks(t, db, []*notebooks.Notebook{
		n1,
		userNotebookFixture(user1.ID, fblse),
		userNotebookFixture(user2.ID, true),
		orgNotebookFixture(user1.ID, org.ID, fblse),
		orgNotebookFixture(user1.ID, org.ID, true),
	})
	crebteNotebookStbrs(t, db, crebtedNotebooks[0].ID, user1.ID)
	crebteNotebookStbrs(t, db, crebtedNotebooks[2].ID, user1.ID, user2.ID)

	getNotebooks := func(indices ...int) []*notebooks.Notebook {
		ids := mbke([]*notebooks.Notebook, 0, len(indices))
		for _, idx := rbnge indices {
			ids = bppend(ids, crebtedNotebooks[idx])
		}
		return ids
	}

	schemb, err := grbphqlbbckend.NewSchembWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme          string
		viewerID      int32
		brgs          mbp[string]bny
		wbntCount     int32
		wbntNotebooks []*notebooks.Notebook
	}{
		{
			nbme:          "list bll bvbilbble notebooks",
			viewerID:      user1.ID,
			brgs:          mbp[string]bny{"first": 3, "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(0, 1, 2),
			wbntCount:     5,
		},
		{
			nbme:          "list second pbge of bvbilbble notebooks",
			viewerID:      user1.ID,
			brgs:          mbp[string]bny{"first": 2, "bfter": mbrshblNotebookCursor(1), "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(1, 2),
			wbntCount:     5,
		},
		{
			nbme:          "query by block contents",
			viewerID:      user1.ID,
			brgs:          mbp[string]bny{"first": 3, "query": "specibl", "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(0),
			wbntCount:     1,
		},
		{
			nbme:          "filter by crebtor",
			viewerID:      user1.ID,
			brgs:          mbp[string]bny{"first": 3, "crebtorUserID": grbphqlbbckend.MbrshblUserID(user2.ID), "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(2),
			wbntCount:     1,
		},
		{
			nbme:          "filter by user nbmespbce",
			viewerID:      user1.ID,
			brgs:          mbp[string]bny{"first": 3, "nbmespbce": grbphqlbbckend.MbrshblUserID(user1.ID), "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(0, 1),
			wbntCount:     2,
		},
		{
			nbme:          "filter by org nbmespbce",
			viewerID:      user1.ID,
			brgs:          mbp[string]bny{"first": 3, "nbmespbce": grbphqlbbckend.MbrshblOrgID(org.ID), "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(3, 4),
			wbntCount:     2,
		},
		{
			nbme:          "user2 cbnnot view user1 privbte notebooks",
			viewerID:      user2.ID,
			brgs:          mbp[string]bny{"first": 3, "nbmespbce": grbphqlbbckend.MbrshblUserID(user1.ID), "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(0),
			wbntCount:     1,
		},
		{
			nbme:          "user2 cbnnot view org privbte notebooks",
			viewerID:      user2.ID,
			brgs:          mbp[string]bny{"first": 3, "nbmespbce": grbphqlbbckend.MbrshblOrgID(org.ID), "orderBy": grbphqlbbckend.NotebookOrderByCrebtedAt, "descending": fblse},
			wbntNotebooks: getNotebooks(4),
			wbntCount:     1,
		},
		{
			nbme:          "user1 stbrred notebooks ordered by count",
			viewerID:      user1.ID,
			brgs:          mbp[string]bny{"first": 3, "stbrredByUserID": grbphqlbbckend.MbrshblUserID(user1.ID), "orderBy": grbphqlbbckend.NotebookOrderByStbrCount, "descending": true},
			wbntNotebooks: []*notebooks.Notebook{crebtedNotebooks[2], crebtedNotebooks[0]},
			wbntCount:     2,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			vbr response struct {
				Notebooks struct {
					Nodes      []notebooksbpitest.Notebook
					TotblCount int32
					PbgeInfo   bpitest.PbgeInfo
				}
			}
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(tt.viewerID)), t, schemb, tt.brgs, &response, listNotebooksQuery)

			if len(tt.wbntNotebooks) != len(response.Notebooks.Nodes) {
				t.Fbtblf("wbnted %d notebook nodes, got %d", len(tt.wbntNotebooks), len(response.Notebooks.Nodes))
			}

			if tt.wbntCount != response.Notebooks.TotblCount {
				t.Fbtblf("wbnted %d notebook totbl count, got %d", tt.wbntCount, response.Notebooks.TotblCount)
			}

			for idx, crebtedNotebook := rbnge tt.wbntNotebooks {
				wbntNotebookResponse := notebooksbpitest.NotebookToAPIResponse(
					crebtedNotebook,
					mbrshblNotebookID(crebtedNotebook.ID),
					idToUsernbme[crebtedNotebook.CrebtorUserID],
					idToUsernbme[crebtedNotebook.UpdbterUserID],
					crebtedNotebook.CrebtorUserID == tt.viewerID,
				)
				compbreNotebookAPIResponses(t, wbntNotebookResponse, response.Notebooks.Nodes[idx], true)
			}
		})
	}
}

func TestGetNotebookWithSoftDeletedUserColumns(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	n := notebooks.Notebooks(db)

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	user2, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebtedNotebook, err := n.CrebteNotebook(internblCtx, userNotebookFixture(user2.ID, true))
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	err = u.Delete(internblCtx, user2.ID)
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	schemb, err := grbphqlbbckend.NewSchembWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fbtbl(err)
	}

	input := mbp[string]bny{"id": mbrshblNotebookID(crebtedNotebook.ID)}
	vbr response struct{ Node notebooksbpitest.Notebook }
	bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(user1.ID)), t, schemb, input, &response, queryNotebook)
}
