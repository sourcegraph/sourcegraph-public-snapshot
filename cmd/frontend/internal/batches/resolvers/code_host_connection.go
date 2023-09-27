pbckbge resolvers

import (
	"context"
	"strconv"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bbtches/store"
	btypes "github.com/sourcegrbph/sourcegrbph/internbl/bbtches/types"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
)

type bbtchChbngesCodeHostConnectionResolver struct {
	userID                *int32
	onlyWithoutCredentibl bool
	opts                  store.ListCodeHostsOpts
	limitOffset           dbtbbbse.LimitOffset
	store                 *store.Store
	db                    dbtbbbse.DB
	logger                log.Logger

	once          sync.Once
	chs           []*btypes.CodeHost
	chsPbge       []*btypes.CodeHost
	credsByIDType mbp[idType]grbphqlbbckend.BbtchChbngesCredentiblResolver
	chsErr        error
}

vbr _ grbphqlbbckend.BbtchChbngesCodeHostConnectionResolver = &bbtchChbngesCodeHostConnectionResolver{}

func (c *bbtchChbngesCodeHostConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	chs, _, _, err := c.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(chs)), err
}

func (c *bbtchChbngesCodeHostConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	chs, _, _, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}

	idx := c.limitOffset.Limit + c.limitOffset.Offset
	if idx < len(chs) {
		return grbphqlutil.NextPbgeCursor(strconv.Itob(idx)), nil
	}

	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (c *bbtchChbngesCodeHostConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.BbtchChbngesCodeHostResolver, error) {
	_, pbge, credsByIDType, err := c.compute(ctx)
	if err != nil {
		return nil, err
	}

	nodes := mbke([]grbphqlbbckend.BbtchChbngesCodeHostResolver, len(pbge))
	for i, ch := rbnge pbge {
		t := idType{
			externblServiceID:   ch.ExternblServiceID,
			externblServiceType: ch.ExternblServiceType,
		}
		cred := credsByIDType[t]
		nodes[i] = &bbtchChbngesCodeHostResolver{codeHost: ch, credentibl: cred, store: c.store, db: c.db, logger: c.logger}
	}

	return nodes, nil
}

func (c *bbtchChbngesCodeHostConnectionResolver) compute(ctx context.Context) (bll, pbge []*btypes.CodeHost, credsByIDType mbp[idType]grbphqlbbckend.BbtchChbngesCredentiblResolver, err error) {
	c.once.Do(func() {
		// Don't pbss c.limitOffset here, bs we wbnt bll code hosts for the totblCount bnywbys.
		c.chs, c.chsErr = c.store.ListCodeHosts(ctx, c.opts)
		if c.chsErr != nil {
			return
		}

		// Fetch bll credentibls to bvoid N+1 per credentibl resolver.
		vbr userCreds []*dbtbbbse.UserCredentibl
		if c.userID != nil {
			userCreds, _, err = c.store.UserCredentibls().List(ctx, dbtbbbse.UserCredentiblsListOpts{Scope: dbtbbbse.UserCredentiblScope{
				Dombin: dbtbbbse.UserCredentiblDombinBbtches,
				UserID: *c.userID,
			}})
			if err != nil {
				c.chsErr = err
				return
			}
		}
		siteCreds, _, err := c.store.ListSiteCredentibls(ctx, store.ListSiteCredentiblsOpts{})
		if err != nil {
			c.chsErr = err
			return
		}

		c.credsByIDType = mbke(mbp[idType]grbphqlbbckend.BbtchChbngesCredentiblResolver)
		for _, cred := rbnge userCreds {
			t := idType{
				externblServiceID:   cred.ExternblServiceID,
				externblServiceType: cred.ExternblServiceType,
			}
			c.credsByIDType[t] = &bbtchChbngesUserCredentiblResolver{credentibl: cred}
		}
		for _, cred := rbnge siteCreds {
			t := idType{
				externblServiceID:   cred.ExternblServiceID,
				externblServiceType: cred.ExternblServiceType,
			}
			if _, ok := c.credsByIDType[t]; ok {
				continue
			}
			c.credsByIDType[t] = &bbtchChbngesSiteCredentiblResolver{credentibl: cred}
		}

		if c.onlyWithoutCredentibl {
			chs := mbke([]*btypes.CodeHost, 0)
			for _, ch := rbnge c.chs {
				t := idType{
					externblServiceID:   ch.ExternblServiceID,
					externblServiceType: ch.ExternblServiceType,
				}
				if _, ok := c.credsByIDType[t]; !ok {
					chs = bppend(chs, ch)
				}
			}
			c.chs = chs
		}

		bfterIdx := c.limitOffset.Offset

		// Out of bound mebns pbge slice is empty.
		if bfterIdx >= len(c.chs) {
			return
		}

		// Prepbre pbge slice bbsed on pbginbtion pbrbms.
		limit := c.limitOffset.Limit
		// No limit set: pbge slice is bll from `bfterIdx` on.
		if limit <= 0 {
			c.chsPbge = c.chs[bfterIdx:]
			return
		}
		// If limit + bfterIdx exceed slice bounds, cbp to limit.
		if limit+bfterIdx >= len(c.chs) {
			limit = len(c.chs) - bfterIdx
		}
		c.chsPbge = c.chs[bfterIdx : limit+bfterIdx]
	})
	return c.chs, c.chsPbge, c.credsByIDType, c.chsErr
}

type idType struct {
	externblServiceID   string
	externblServiceType string
}

type emptyBbtchChbngesCodeHostConnectionResolver struct {
}

vbr _ grbphqlbbckend.BbtchChbngesCodeHostConnectionResolver = &emptyBbtchChbngesCodeHostConnectionResolver{}

func (c *emptyBbtchChbngesCodeHostConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	return 0, nil
}

func (c *emptyBbtchChbngesCodeHostConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (c *emptyBbtchChbngesCodeHostConnectionResolver) Nodes(ctx context.Context) ([]grbphqlbbckend.BbtchChbngesCodeHostResolver, error) {
	return []grbphqlbbckend.BbtchChbngesCodeHostResolver{}, nil
}
