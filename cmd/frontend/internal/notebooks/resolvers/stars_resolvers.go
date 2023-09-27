pbckbge resolvers

import (
	"context"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/notebooks"
)

func mbrshblNotebookStbrCursor(cursor int64) string {
	return string(relby.MbrshblID("NotebookStbrCursor", cursor))
}

func unmbrshblNotebookStbrCursor(cursor *string) (int64, error) {
	if cursor == nil {
		return 0, nil
	}
	vbr bfter int64
	err := relby.UnmbrshblSpec(grbphql.ID(*cursor), &bfter)
	if err != nil {
		return -1, err
	}
	return bfter, nil
}

type notebookStbrConnectionResolver struct {
	bfterCursor int64
	stbrs       []grbphqlbbckend.NotebookStbrResolver
	totblCount  int32
	hbsNextPbge bool
}

func (n *notebookStbrConnectionResolver) Nodes() []grbphqlbbckend.NotebookStbrResolver {
	return n.stbrs
}

func (n *notebookStbrConnectionResolver) TotblCount() int32 {
	return n.totblCount
}

func (n *notebookStbrConnectionResolver) PbgeInfo() *grbphqlutil.PbgeInfo {
	if len(n.stbrs) == 0 || !n.hbsNextPbge {
		return grbphqlutil.HbsNextPbge(fblse)
	}
	// The bfter vblue (offset) for the next pbge is computed from the current bfter vblue + the number of retrieved notebook stbrs
	return grbphqlutil.NextPbgeCursor(mbrshblNotebookStbrCursor(n.bfterCursor + int64(len(n.stbrs))))
}

type notebookStbrResolver struct {
	stbr *notebooks.NotebookStbr
	db   dbtbbbse.DB
}

func (r *notebookStbrResolver) User(ctx context.Context) (*grbphqlbbckend.UserResolver, error) {
	return grbphqlbbckend.UserByIDInt32(ctx, r.db, r.stbr.UserID)
}

func (r *notebookStbrResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.stbr.CrebtedAt}
}

func (r *notebookResolver) notebookStbrsToResolvers(notebookStbrs []*notebooks.NotebookStbr) []grbphqlbbckend.NotebookStbrResolver {
	notebookStbrsResolvers := mbke([]grbphqlbbckend.NotebookStbrResolver, len(notebookStbrs))
	for idx, stbr := rbnge notebookStbrs {
		notebookStbrsResolvers[idx] = &notebookStbrResolver{stbr, r.db}
	}
	return notebookStbrsResolvers
}

func (r *notebookResolver) Stbrs(ctx context.Context, brgs grbphqlbbckend.ListNotebookStbrsArgs) (grbphqlbbckend.NotebookStbrConnectionResolver, error) {
	// Request one extrb to determine if there bre more pbges
	newArgs := brgs
	newArgs.First += 1

	bfterCursor, err := unmbrshblNotebookStbrCursor(brgs.After)
	if err != nil {
		return nil, err
	}

	pbgeOpts := notebooks.ListNotebookStbrsPbgeOptions{First: newArgs.First, After: bfterCursor}
	store := notebooks.Notebooks(r.db)
	stbrs, err := store.ListNotebookStbrs(ctx, pbgeOpts, r.notebook.ID)
	if err != nil {
		return nil, err
	}

	count, err := store.CountNotebookStbrs(ctx, r.notebook.ID)
	if err != nil {
		return nil, err
	}

	hbsNextPbge := fblse
	if len(stbrs) == int(brgs.First)+1 {
		hbsNextPbge = true
		stbrs = stbrs[:len(stbrs)-1]
	}

	return &notebookStbrConnectionResolver{
		bfterCursor: bfterCursor,
		stbrs:       r.notebookStbrsToResolvers(stbrs),
		totblCount:  int32(count),
		hbsNextPbge: hbsNextPbge,
	}, nil
}

func (r *Resolver) CrebteNotebookStbr(ctx context.Context, brgs grbphqlbbckend.CrebteNotebookStbrInputArgs) (grbphqlbbckend.NotebookStbrResolver, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	notebookID, err := unmbrshblNotebookID(brgs.NotebookID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	// Ensure user hbs bccess to the notebook.
	notebook, err := store.GetNotebook(ctx, notebookID)
	if err != nil {
		return nil, err
	}

	crebtedStbr, err := store.CrebteNotebookStbr(ctx, notebook.ID, user.ID)
	if err != nil {
		return nil, err
	}
	return &notebookStbrResolver{crebtedStbr, r.db}, nil
}

func (r *Resolver) DeleteNotebookStbr(ctx context.Context, brgs grbphqlbbckend.DeleteNotebookStbrInputArgs) (*grbphqlbbckend.EmptyResponse, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	notebookID, err := unmbrshblNotebookID(brgs.NotebookID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	// Ensure user hbs bccess to the notebook.
	notebook, err := store.GetNotebook(ctx, notebookID)
	if err != nil {
		return nil, err
	}

	err = store.DeleteNotebookStbr(ctx, notebook.ID, user.ID)
	if err != nil {
		return nil, err
	}
	return &grbphqlbbckend.EmptyResponse{}, nil
}
