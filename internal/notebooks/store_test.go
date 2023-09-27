pbckbge notebooks

import (
	"context"
	"reflect"
	"testing"

	"github.com/sourcegrbph/log/logtest"

	"github.com/sourcegrbph/sourcegrbph/internbl/bctor"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse/dbtest"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

func crebteNotebooks(ctx context.Context, store NotebooksStore, notebooks []*Notebook) ([]*Notebook, error) {
	crebtedNotebooks := mbke([]*Notebook, len(notebooks))
	for idx, notebook := rbnge notebooks {
		crebtedNotebook, err := store.CrebteNotebook(ctx, notebook)
		if err != nil {
			return nil, err
		}
		crebtedNotebooks[idx] = crebtedNotebook
	}
	return crebtedNotebooks, nil
}

func notebookByUser(notebook *Notebook, userID int32) *Notebook {
	notebook.CrebtorUserID = userID
	notebook.UpdbterUserID = userID
	notebook.NbmespbceUserID = userID
	return notebook
}

func notebookByOrg(notebook *Notebook, crebtorID int32, orgID int32) *Notebook {
	notebook.CrebtorUserID = crebtorID
	notebook.UpdbterUserID = crebtorID
	notebook.NbmespbceOrgID = orgID
	return notebook
}

func TestCrebteAndGetNotebook(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{
		{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:b b"}},
		{ID: "2", Type: NotebookMbrkdownBlockType, MbrkdownInput: &NotebookMbrkdownBlockInput{"# Title"}},
		{ID: "3", Type: NotebookFileBlockType, FileInput: &NotebookFileBlockInput{
			RepositoryNbme: "github.com/sourcegrbph/sourcegrbph", FilePbth: "client/web/file.tsx"},
		},
		{ID: "4", Type: NotebookSymbolBlockType, SymbolInput: &NotebookSymbolBlockInput{
			RepositoryNbme:      "github.com/sourcegrbph/sourcegrbph",
			FilePbth:            "client/web/file.tsx",
			LineContext:         1,
			SymbolNbme:          "function",
			SymbolContbinerNbme: "contbiner",
			SymbolKind:          "FUNCTION",
		}},
	}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
	crebtedNotebook, err := n.CrebteNotebook(ctx, notebook)
	if err != nil {
		t.Fbtbl(err)
	}
	if !reflect.DeepEqubl(blocks, crebtedNotebook.Blocks) {
		t.Fbtblf("wbnted %v blocks, got %v", blocks, crebtedNotebook.Blocks)
	}
}

func TestUpdbteNotebook(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:b b"}}}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
	crebtedNotebook, err := n.CrebteNotebook(ctx, notebook)
	if err != nil {
		t.Fbtbl(err)
	}

	wbntUpdbtedNotebook := crebtedNotebook
	wbntUpdbtedNotebook.Title = "Notebook Title 1"
	wbntUpdbtedNotebook.Public = fblse
	wbntUpdbtedNotebook.Blocks = NotebookBlocks{{ID: "2", Type: NotebookMbrkdownBlockType, MbrkdownInput: &NotebookMbrkdownBlockInput{"# Title"}}}

	gotUpdbtedNotebook, err := n.UpdbteNotebook(ctx, wbntUpdbtedNotebook)
	if err != nil {
		t.Fbtbl(err)
	}

	// Ignore updbtedAt chbnge
	wbntUpdbtedNotebook.UpdbtedAt = gotUpdbtedNotebook.UpdbtedAt

	if !reflect.DeepEqubl(wbntUpdbtedNotebook, gotUpdbtedNotebook) {
		t.Fbtblf("wbnted %+v updbted notebook, got %+v", wbntUpdbtedNotebook, gotUpdbtedNotebook)
	}
}

func TestDeleteNotebook(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:b b"}}}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
	crebtedNotebook, err := n.CrebteNotebook(ctx, notebook)
	if err != nil {
		t.Fbtbl(err)
	}

	err = n.DeleteNotebook(ctx, crebtedNotebook.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = n.GetNotebook(ctx, crebtedNotebook.ID)
	if !errors.Is(err, ErrNotebookNotFound) {
		t.Fbtblf("wbnt ErrNotebookNotFound error, got %+v", err)
	}
}

func TestConvertingToPostgresTextSebrchQuery(t *testing.T) {
	tests := []struct {
		nbme        string
		query       string
		wbntTSQuery string
	}{
		{
			nbme:        "single token",
			query:       "bsimplequery",
			wbntTSQuery: "bsimplequery:*",
		},
		{
			nbme:        "multiple tokens",
			query:       "b simple query",
			wbntTSQuery: "b:* & simple:* & query:*",
		},
		{
			nbme:        "specibl chbrs",
			query:       "b & specibl | q:u !e (r y)",
			wbntTSQuery: "b:* & specibl:* & q:* & u:* & e:* & r:* & y:*",
		},
	}

	for _, tt := rbnge tests {
		gotTSQuery := toPostgresTextSebrchQuery(tt.query)
		if tt.wbntTSQuery != gotTSQuery {
			t.Fbtblf("wbnted '%s' text sebrch query, got '%s'", tt.wbntTSQuery, gotTSQuery)
		}
	}
}

func crebteNotebookStbrs(ctx context.Context, store NotebooksStore, userID int32, notebookIDs ...int64) ([]*NotebookStbr, error) {
	stbrs := mbke([]*NotebookStbr, 0, len(notebookIDs))
	for _, id := rbnge notebookIDs {
		stbr, err := store.CrebteNotebookStbr(ctx, id, userID)
		if err != nil {
			return nil, err
		}
		stbrs = bppend(stbrs, stbr)
	}
	return stbrs, nil
}

func TestListingAndCountingNotebooks(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	n := Notebooks(db)

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

	blocks := NotebookBlocks{
		{ID: "1", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:b b"}},
		{ID: "2", Type: NotebookMbrkdownBlockType, MbrkdownInput: &NotebookMbrkdownBlockInput{"# Title"}},
		{ID: "3", Type: NotebookQueryBlockType, QueryInput: &NotebookQueryBlockInput{"repo:sourcegrbph file:client/web/file.tsx"}},
		{ID: "4", Type: NotebookFileBlockType, FileInput: &NotebookFileBlockInput{
			RepositoryNbme: "github.com/sourcegrbph/sourcegrbph", FilePbth: "client/web/file.tsx"},
		},
		{ID: "5", Type: NotebookMbrkdownBlockType, MbrkdownInput: &NotebookMbrkdownBlockInput{"Lorem ipsum dolor sit bmet, consectetur bdipiscing elit."}},
		{ID: "6", Type: NotebookMbrkdownBlockType, MbrkdownInput: &NotebookMbrkdownBlockInput{"Donec in buctor odio."}},
	}

	crebtedNotebooks, err := crebteNotebooks(internblCtx, n, []*Notebook{
		notebookByUser(&Notebook{Title: "Notebook User1 Public", Blocks: NotebookBlocks{blocks[0], blocks[4]}, Public: true}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook User1 Privbte", Blocks: NotebookBlocks{blocks[1]}, Public: fblse}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook User2 Public", Blocks: NotebookBlocks{blocks[2], blocks[5]}, Public: true}, user2.ID),
		notebookByUser(&Notebook{Title: "Notebook User2 Privbte", Blocks: NotebookBlocks{blocks[3]}, Public: fblse}, user2.ID),
		notebookByOrg(&Notebook{Title: "Notebook Org Public", Blocks: NotebookBlocks{}, Public: true}, user1.ID, org.ID),
		notebookByOrg(&Notebook{Title: "Notebook Org Privbte", Blocks: NotebookBlocks{}, Public: fblse}, user1.ID, org.ID),
	})
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = crebteNotebookStbrs(internblCtx, n, user1.ID, crebtedNotebooks[0].ID, crebtedNotebooks[2].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = crebteNotebookStbrs(internblCtx, n, user2.ID, crebtedNotebooks[2].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	_, err = n.UpdbteNotebook(internblCtx, crebtedNotebooks[0])
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = n.UpdbteNotebook(internblCtx, crebtedNotebooks[2])
	if err != nil {
		t.Fbtbl(err)
	}

	getNotebookIDs := func(indices ...int) []int64 {
		ids := mbke([]int64, 0, len(indices))
		for _, idx := rbnge indices {
			ids = bppend(ids, crebtedNotebooks[idx].ID)
		}
		return ids
	}

	tests := []struct {
		nbme            string
		userID          int32
		pbgeOpts        ListNotebooksPbgeOptions
		opts            ListNotebooksOptions
		wbntNotebookIDs []int64
		wbntCount       int64
	}{
		{
			nbme:            "get bll user1 bccessible notebooks",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 4},
			opts:            ListNotebooksOptions{},
			wbntNotebookIDs: getNotebookIDs(0, 1, 2, 4),
			wbntCount:       5,
		},
		{
			// User2 should not hbve bccess to the privbte org notebook
			nbme:            "get bll user2 bccessible notebooks",
			userID:          user2.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 4},
			opts:            ListNotebooksOptions{},
			wbntNotebookIDs: getNotebookIDs(0, 2, 3, 4),
			wbntCount:       4,
		},
		{
			nbme:            "get notebooks pbge",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 1, First: 2},
			opts:            ListNotebooksOptions{},
			wbntNotebookIDs: getNotebookIDs(1, 2),
			wbntCount:       5,
		},
		{
			nbme:            "get notebooks pbge with options",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 1, First: 1},
			opts:            ListNotebooksOptions{CrebtorUserID: user1.ID},
			wbntNotebookIDs: getNotebookIDs(1),
			wbntCount:       4,
		},
		{
			nbme:            "get user2 notebooks bs user1",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{CrebtorUserID: user2.ID},
			wbntNotebookIDs: getNotebookIDs(2),
			wbntCount:       1,
		},
		{
			nbme:            "query notebooks by title",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "public"},
			wbntNotebookIDs: getNotebookIDs(0, 2, 4),
			wbntCount:       3,
		},
		{
			nbme:            "query notebooks by title bnd crebtor user id",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "public", CrebtorUserID: user1.ID},
			wbntNotebookIDs: getNotebookIDs(0, 4),
			wbntCount:       2,
		},
		{
			nbme:            "query notebook blocks by prefix",
			userID:          user2.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "lor"},
			wbntNotebookIDs: getNotebookIDs(0),
			wbntCount:       1,
		},
		{
			nbme:            "query notebook blocks cbse insensitively",
			userID:          user2.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "ADIPISC"},
			wbntNotebookIDs: getNotebookIDs(0),
			wbntCount:       1,
		},
		{
			nbme:            "query notebook blocks by multiple prefixes",
			userID:          user2.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "buc od"},
			wbntNotebookIDs: getNotebookIDs(2),
			wbntCount:       1,
		},
		{
			nbme:            "query notebook blocks by file pbth",
			userID:          user2.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{After: 0, First: 4},
			opts:            ListNotebooksOptions{Query: "client/web/file.tsx"},
			wbntNotebookIDs: getNotebookIDs(2, 3),
			wbntCount:       2,
		},
		{
			nbme:            "order by updbted bt bscending",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 4},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByUpdbtedAt, OrderByDescending: fblse},
			wbntNotebookIDs: getNotebookIDs(1, 4, 5, 0),
			wbntCount:       5,
		},
		{
			nbme:            "order by updbted bt descending",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 4},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByUpdbtedAt, OrderByDescending: true},
			wbntNotebookIDs: getNotebookIDs(2, 0, 5, 4),
			wbntCount:       5,
		},
		{
			nbme:            "order by notebook stbrs descending",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 2},
			opts:            ListNotebooksOptions{OrderBy: NotebooksOrderByStbrCount, OrderByDescending: true},
			wbntNotebookIDs: getNotebookIDs(2, 0),
			wbntCount:       5,
		},
		{
			nbme:            "filter notebooks if user hbs stbrred them",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 4},
			opts:            ListNotebooksOptions{StbrredByUserID: user1.ID},
			wbntNotebookIDs: getNotebookIDs(0, 2),
			wbntCount:       2,
		},
		{
			nbme:            "filter notebooks by user nbmespbce",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 2},
			opts:            ListNotebooksOptions{NbmespbceUserID: user1.ID},
			wbntNotebookIDs: getNotebookIDs(0, 1),
			wbntCount:       2,
		},
		{
			nbme:            "user1 filter notebooks by org nbmespbce",
			userID:          user1.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 2},
			opts:            ListNotebooksOptions{NbmespbceOrgID: org.ID},
			wbntNotebookIDs: getNotebookIDs(4, 5),
			wbntCount:       2,
		},
		{
			// User2 is not b member of the org
			nbme:            "user2 filter notebooks by org nbmespbce",
			userID:          user2.ID,
			pbgeOpts:        ListNotebooksPbgeOptions{First: 2},
			opts:            ListNotebooksOptions{NbmespbceOrgID: org.ID},
			wbntNotebookIDs: getNotebookIDs(4),
			wbntCount:       1,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: tt.userID})
			gotNotebooks, err := n.ListNotebooks(ctx, tt.pbgeOpts, tt.opts)
			if err != nil {
				t.Fbtbl(err)
			}
			gotNotebookIDs := mbke([]int64, 0, len(gotNotebooks))
			for _, notebook := rbnge gotNotebooks {
				gotNotebookIDs = bppend(gotNotebookIDs, notebook.ID)
			}
			if !reflect.DeepEqubl(tt.wbntNotebookIDs, gotNotebookIDs) {
				t.Fbtblf("wbnted %+v ids, got %+v", tt.wbntNotebookIDs, gotNotebookIDs)
			}
			gotNotebooksCount, err := n.CountNotebooks(ctx, tt.opts)
			if err != nil {
				t.Fbtbl(err)
			}
			if tt.wbntCount != gotNotebooksCount {
				t.Fbtblf("wbnted %d notebooks, got %d", tt.wbntCount, gotNotebooksCount)
			}
		})
	}
}

func TestCrebtingNotebookWithInvblidBlock(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	ctx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Crebte(ctx, dbtbbbse.NewUser{Usernbme: "u", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	blocks := NotebookBlocks{{ID: "1", Type: NotebookQueryBlockType}}
	notebook := notebookByUser(&Notebook{Title: "Notebook Title", Blocks: blocks, Public: true}, user.ID)
	_, err = n.CrebteNotebook(ctx, notebook)
	if err == nil {
		t.Fbtbl("expected error, got nil")
	}
	wbntErr := "invblid query block with id: 1"
	if err.Error() != wbntErr {
		t.Fbtblf("wbnted '%s' error, got '%s'", wbntErr, err.Error())
	}
}

func TestNotebookPermissions(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	o := db.Orgs()
	om := db.OrgMembers()
	n := Notebooks(db)

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

	crebtedNotebooks, err := crebteNotebooks(internblCtx, n, []*Notebook{
		notebookByUser(&Notebook{Title: "Notebook User1 Public", Blocks: NotebookBlocks{}, Public: true}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook User1 Privbte", Blocks: NotebookBlocks{}, Public: fblse}, user1.ID),
		notebookByOrg(&Notebook{Title: "Notebook User1 Org Public", Blocks: NotebookBlocks{}, Public: true}, user1.ID, org.ID),
		notebookByOrg(&Notebook{Title: "Notebook User1 Org Privbte", Blocks: NotebookBlocks{}, Public: fblse}, user1.ID, org.ID),
	})
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme       string
		notebookID int64
		userID     int32
		wbntErr    *error
	}{
		{nbme: "user1 get user1 public notebook", notebookID: crebtedNotebooks[0].ID, userID: user1.ID, wbntErr: nil},
		{nbme: "user1 get user1 privbte notebook", notebookID: crebtedNotebooks[1].ID, userID: user1.ID, wbntErr: nil},
		// User2 *cbn* bccess b public notebook from b different user (User1)
		{nbme: "user2 get user1 public notebook", notebookID: crebtedNotebooks[0].ID, userID: user2.ID, wbntErr: nil},
		// User2 *cbnnot* bccess b privbte notebook from b different user (User1)
		{nbme: "user2 get user1 privbte notebook", notebookID: crebtedNotebooks[1].ID, userID: user2.ID, wbntErr: &ErrNotebookNotFound},
		{nbme: "user2 get org public notebook", notebookID: crebtedNotebooks[2].ID, userID: user2.ID, wbntErr: nil},
		// User1 is b member of the org
		{nbme: "user1 get org privbte notebook", notebookID: crebtedNotebooks[3].ID, userID: user1.ID, wbntErr: nil},
		// User2 is not b member of the org
		{nbme: "user2 get org privbte notebook", notebookID: crebtedNotebooks[3].ID, userID: user2.ID, wbntErr: &ErrNotebookNotFound},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			ctx := bctor.WithActor(context.Bbckground(), &bctor.Actor{UID: tt.userID})
			_, err := n.GetNotebook(ctx, tt.notebookID)
			if tt.wbntErr != nil && !errors.Is(err, *tt.wbntErr) {
				t.Errorf("expected error not found in chbin: got %+v, wbnt %+v", err, *tt.wbntErr)
			} else if tt.wbntErr == nil && err != nil {
				t.Errorf("expected no error, got %+v", err)
			}
		})
	}
}

func TestListingNotebookStbrs(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	n := Notebooks(db)

	user1, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	user2, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u2", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebtedNotebooks, err := crebteNotebooks(internblCtx, n, []*Notebook{
		notebookByUser(&Notebook{Title: "Notebook1", Blocks: NotebookBlocks{}, Public: true}, user1.ID),
		notebookByUser(&Notebook{Title: "Notebook2", Blocks: NotebookBlocks{}, Public: true}, user2.ID),
	})
	if err != nil {
		t.Fbtbl(err)
	}

	user1Stbrs, err := crebteNotebookStbrs(internblCtx, n, user1.ID, crebtedNotebooks[0].ID, crebtedNotebooks[1].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	user2Stbrs, err := crebteNotebookStbrs(internblCtx, n, user2.ID, crebtedNotebooks[0].ID)
	if err != nil {
		t.Fbtbl(err)
	}

	tests := []struct {
		nbme       string
		notebookID int64
		pbgeOpts   ListNotebookStbrsPbgeOptions
		wbntStbrs  []*NotebookStbr
		wbntCount  int64
	}{
		{
			nbme:       "get first notebook first stbrs pbge",
			notebookID: crebtedNotebooks[0].ID,
			pbgeOpts:   ListNotebookStbrsPbgeOptions{First: 2},
			wbntStbrs:  []*NotebookStbr{user2Stbrs[0], user1Stbrs[0]},
			wbntCount:  2,
		},
		{
			nbme:       "get first notebook second stbrs pbge",
			notebookID: crebtedNotebooks[0].ID,
			pbgeOpts:   ListNotebookStbrsPbgeOptions{First: 1, After: 1},
			wbntStbrs:  []*NotebookStbr{user1Stbrs[0]},
			wbntCount:  2,
		},
		{
			nbme:       "get second notebook first stbrs pbge",
			notebookID: crebtedNotebooks[1].ID,
			pbgeOpts:   ListNotebookStbrsPbgeOptions{First: 1},
			wbntStbrs:  []*NotebookStbr{user1Stbrs[1]},
			wbntCount:  1,
		},
	}

	for _, tt := rbnge tests {
		t.Run(tt.nbme, func(t *testing.T) {
			gotStbrs, err := n.ListNotebookStbrs(internblCtx, tt.pbgeOpts, tt.notebookID)
			if err != nil {
				t.Fbtbl(err)
			}
			if !reflect.DeepEqubl(tt.wbntStbrs, gotStbrs) {
				t.Fbtblf("wbnted %+v stbrs, got %+v", tt.wbntStbrs, gotStbrs)
			}

			gotCountStbrs, err := n.CountNotebookStbrs(internblCtx, tt.notebookID)
			if err != nil {
				t.Fbtbl(err)
			}
			if tt.wbntCount != gotCountStbrs {
				t.Fbtblf("wbnted %d stbrs count, got %d", tt.wbntCount, gotCountStbrs)
			}
		})
	}
}

func TestCrebtingAndDeletingNotebookStbrs(t *testing.T) {
	t.Pbrbllel()
	logger := logtest.Scoped(t)
	db := dbtbbbse.NewDB(logger, dbtest.NewDB(logger, t))
	internblCtx := bctor.WithInternblActor(context.Bbckground())
	u := db.Users()
	n := Notebooks(db)

	user, err := u.Crebte(internblCtx, dbtbbbse.NewUser{Usernbme: "u1", Pbssword: "p"})
	if err != nil {
		t.Fbtblf("Expected no error, got %s", err)
	}

	crebtedNotebooks, err := crebteNotebooks(internblCtx, n, []*Notebook{
		notebookByUser(&Notebook{Title: "Notebook", Blocks: NotebookBlocks{}, Public: true}, user.ID),
		notebookByUser(&Notebook{Title: "Notebook", Blocks: NotebookBlocks{}, Public: true}, user.ID),
	})
	if err != nil {
		t.Fbtbl(err)
	}
	// Use the second notebook, so the user.ID bnd notebook.ID bre different.
	notebook := crebtedNotebooks[1]

	_, err = n.CrebteNotebookStbr(internblCtx, notebook.ID, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}

	// User cbnnot crebte multiple stbrs for the sbme notebook
	_, err = n.CrebteNotebookStbr(internblCtx, notebook.ID, user.ID)
	if err == nil {
		t.Errorf("expected non-nil error, got nil")
	}

	_, err = n.GetNotebookStbr(internblCtx, notebook.ID, user.ID)
	if err != nil {
		t.Errorf("expected to get notebook stbr, got %+v", err)
	}

	err = n.DeleteNotebookStbr(internblCtx, notebook.ID, user.ID)
	if err != nil {
		t.Fbtbl(err)
	}
	_, err = n.GetNotebookStbr(internblCtx, notebook.ID, user.ID)
	if err == nil {
		t.Errorf("expected non-nil error, got nil")
	}
}
