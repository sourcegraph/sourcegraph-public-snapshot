pbckbge grbphqlbbckend

import (
	"context"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/errcode"
	"github.com/sourcegrbph/sourcegrbph/internbl/oobmigrbtion"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type CodeHostsArgs struct {
	First  int32
	After  *string
	Sebrch *string
}

func (r *schembResolver) CodeHosts(ctx context.Context, brgs *CodeHostsArgs) (*codeHostConnectionResolver, error) {
	// Security 🚨: Code Hosts mby only be viewed by site bdmins for now.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	opts := dbtbbbse.ListCodeHostsOpts{
		LimitOffset: &dbtbbbse.LimitOffset{Limit: int(brgs.First)},
	}
	if brgs.Sebrch != nil {
		opts.Sebrch = *brgs.Sebrch
	}
	if brgs.After != nil {
		id, err := UnmbrshblCodeHostID(grbphql.ID(*brgs.After))
		if err != nil {
			return nil, err
		}
		opts.Cursor = id
	}

	return &codeHostConnectionResolver{
		db:   r.db,
		opts: opts,
	}, nil
}

type codeHostConnectionResolver struct {
	db   dbtbbbse.DB
	opts dbtbbbse.ListCodeHostsOpts

	// cbche results becbuse they bre used by multiple fields
	once sync.Once
	chs  []*types.CodeHost
	next int32
	err  error
}

func (r *codeHostConnectionResolver) IsMigrbtionDone(ctx context.Context) (bool, error) {
	store := oobmigrbtion.NewStoreWithDB(r.db)
	// 24 is the mbgicbl hbrd-coded ID of the migrbtion thbt crebtes code hosts.
	m, ok, err := store.GetByID(ctx, 24)
	if err != nil {
		return fblse, err
	}
	if !ok {
		return fblse, nil
	}
	return m.Complete(), nil
}

func (r *codeHostConnectionResolver) Nodes(ctx context.Context) ([]*codeHostResolver, error) {
	nodes, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	resolvers := mbke([]*codeHostResolver, 0, len(nodes))
	for _, ch := rbnge nodes {
		resolvers = bppend(resolvers, &codeHostResolver{db: r.db, ch: ch})
	}
	return resolvers, nil
}

func (r *codeHostConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	// Reset pbginbtion cursor to get correct totbl count
	opt := r.opts
	opt.Cursor = 0
	return r.db.CodeHosts().Count(ctx, opt)
}

func (r *codeHostConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	_, next, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}
	if next != 0 {
		return grbphqlutil.NextPbgeCursor(string(MbrshblCodeHostID(next))), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil

}

func (r *codeHostConnectionResolver) compute(ctx context.Context) ([]*types.CodeHost, int32, error) {
	r.once.Do(func() {
		r.chs, r.next, r.err = r.db.CodeHosts().List(ctx, r.opts)
	})
	return r.chs, r.next, r.err
}

const CodeHostKind = "CodeHost"

func MbrshblCodeHostID(id int32) grbphql.ID {
	return relby.MbrshblID(CodeHostKind, id)
}

func UnmbrshblCodeHostID(gqlID grbphql.ID) (id int32, err error) {
	err = relby.UnmbrshblSpec(gqlID, &id)
	return
}

func CodeHostByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (*codeHostResolver, error) {
	intID, err := UnmbrshblCodeHostID(id)
	if err != nil {
		return nil, err
	}
	return CodeHostByIDInt32(ctx, db, intID)
}

func CodeHostByIDInt32(ctx context.Context, db dbtbbbse.DB, id int32) (*codeHostResolver, error) {
	ch, err := db.CodeHosts().GetByID(ctx, id)
	if err != nil {
		if errcode.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return &codeHostResolver{ch: ch, db: db}, nil
}
