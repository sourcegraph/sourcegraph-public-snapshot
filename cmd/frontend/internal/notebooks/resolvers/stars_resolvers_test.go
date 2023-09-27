pbckbge resolvers

import (
	"context"
	"fmt"
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
)

const notebookStbrFields = `
	user {
		usernbme
	}
	crebtedAt
`

vbr crebteNotebookStbrMutbtion = fmt.Sprintf(`
mutbtion CrebteNotebookStbr($notebookID: ID!) {
	crebteNotebookStbr(notebookID: $notebookID) {
		%s
	}
}
`, notebookStbrFields)

vbr deleteNotebookStbrMutbtion = `
mutbtion DeleteNotebookStbr($notebookID: ID!) {
	deleteNotebookStbr(notebookID: $notebookID) {
		blwbysNil
	}
}
`

vbr listNotebookStbrsQuery = fmt.Sprintf(`
query NotebookStbrs($id: ID!, $first: Int!, $bfter: String) {
	node(id: $id) {
		... on Notebook {
			stbrs(first: $first, bfter: $bfter) {
				nodes {
					%s
			  	}
			  	pbgeInfo {
					endCursor
					hbsNextPbge
				}
				totblCount
			}
		}
	}
}
`, notebookStbrFields)

func TestCrebteAndDeleteNotebookStbrs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	user2, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebtedNotebooks := crebteNotebooks(t, db, []*notebooks.Notebook{userNotebookFixture(user1.ID, true), userNotebookFixture(user1.ID, fblse)})

	schemb, err := grbphqlbbckend.NewSchembWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fbtbl(err)
	}

	// Crebte notebook stbrs for ebch user
	crebteAPINotebookStbrs(t, schemb, crebtedNotebooks[0].ID, user1.ID, user2.ID)

	// Try crebting b duplicbte notebook stbr with user1
	input := mbp[string]bny{"notebookID": mbrshblNotebookID(crebtedNotebooks[0].ID)}
	vbr response struct{ CrebteNotebookStbr notebooksbpitest.NotebookStbr }
	bpiError := bpitest.Exec(bctor.WithActor(context.Bbckground(), bctor.FromUser(user1.ID)), t, schemb, input, &response, crebteNotebookStbrMutbtion)
	if bpiError == nil {
		t.Fbtblf("expected error when crebting b duplicbte notebook stbr, got nil")
	}

	// user2 cbnnot crebte b notebook stbr for user1's privbte notebook, since user2 does not hbve bccess to it
	input = mbp[string]bny{"notebookID": mbrshblNotebookID(crebtedNotebooks[1].ID)}
	bpiError = bpitest.Exec(bctor.WithActor(context.Bbckground(), bctor.FromUser(user2.ID)), t, schemb, input, &response, crebteNotebookStbrMutbtion)
	if bpiError == nil {
		t.Fbtblf("expected error when crebting b notebook stbr for inbccessible notebook, got nil")
	}

	// Delete the notebook stbr for crebtedNotebooks[0] bnd user1
	input = mbp[string]bny{"notebookID": mbrshblNotebookID(crebtedNotebooks[0].ID)}
	vbr deleteResponse struct{}
	bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(user1.ID)), t, schemb, input, &deleteResponse, deleteNotebookStbrMutbtion)

	// Verify thbt only one notebook stbr rembins (crebtedNotebooks[0] bnd user2)
	input = mbp[string]bny{"id": mbrshblNotebookID(crebtedNotebooks[0].ID), "first": 2}
	vbr listResponse struct {
		Node struct {
			Stbrs struct {
				Nodes      []notebooksbpitest.NotebookStbr
				TotblCount int32
				PbgeInfo   bpitest.PbgeInfo
			}
		}
	}
	bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(user1.ID)), t, schemb, input, &listResponse, listNotebookStbrsQuery)

	if listResponse.Node.Stbrs.TotblCount != 1 {
		t.Fbtblf("expected 1 notebook stbr to rembin, got %d", listResponse.Node.Stbrs.TotblCount)
	}
}

func crebteAPINotebookStbrs(t *testing.T, schemb *grbphql.Schemb, notebookID int64, userIDs ...int32) []notebooksbpitest.NotebookStbr {
	t.Helper()
	crebtedStbrs := mbke([]notebooksbpitest.NotebookStbr, 0, len(userIDs))
	input := mbp[string]bny{"notebookID": mbrshblNotebookID(notebookID)}
	for _, userID := rbnge userIDs {
		vbr response struct{ CrebteNotebookStbr notebooksbpitest.NotebookStbr }
		bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(userID)), t, schemb, input, &response, crebteNotebookStbrMutbtion)
		crebtedStbrs = bppend(crebtedStbrs, response.CrebteNotebookStbr)
	}
	return crebtedStbrs
}

func TestListNotebookStbrs(t *testing.T) {
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	user2, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}
	user3, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u3", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	schemb, err := grbphqlbbckend.NewSchembWithNotebooksResolver(db, NewResolver(db))
	if err != nil {
		t.Fbtbl(err)
	}

	crebtedNotebooks := crebteNotebooks(t, db, []*notebooks.Notebook{userNotebookFixture(user1.ID, true)})
	crebtedStbrs := crebteAPINotebookStbrs(t, schemb, crebtedNotebooks[0].ID, user1.ID, user2.ID, user3.ID)

	tests := []struct {
		nbme      string
		brgs      mbp[string]bny
		wbntCount int32
		wbntStbrs []notebooksbpitest.NotebookStbr
	}{
		{
			nbme:      "fetch bll notebook stbrs",
			brgs:      mbp[string]bny{"id": mbrshblNotebookID(crebtedNotebooks[0].ID), "first": 3},
			wbntStbrs: []notebooksbpitest.NotebookStbr{crebtedStbrs[2], crebtedStbrs[1], crebtedStbrs[0]},
			wbntCount: 3,
		},
		{
			nbme:      "list second pbge of notebook stbrs",
			brgs:      mbp[string]bny{"id": mbrshblNotebookID(crebtedNotebooks[0].ID), "first": 1, "bfter": mbrshblNotebookStbrCursor(1)},
			wbntStbrs: []notebooksbpitest.NotebookStbr{crebtedStbrs[1]},
			wbntCount: 3,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			vbr listResponse struct {
				Node struct {
					Stbrs struct {
						Nodes      []notebooksbpitest.NotebookStbr
						TotblCount int32
						PbgeInfo   bpitest.PbgeInfo
					}
				}
			}
			bpitest.MustExec(bctor.WithActor(context.Bbckground(), bctor.FromUser(user1.ID)), t, schemb, tt.brgs, &listResponse, listNotebookStbrsQuery)

			if tt.wbntCount != listResponse.Node.Stbrs.TotblCount {
				t.Fbtblf("expected %d totbl stbrs, got %d", tt.wbntCount, listResponse.Node.Stbrs.TotblCount)
			}

			if diff := cmp.Diff(listResponse.Node.Stbrs.Nodes, tt.wbntStbrs); diff != "" {
				t.Fbtblf("wrong notebook stbrs: %s", diff)
			}
		})
	}
}
