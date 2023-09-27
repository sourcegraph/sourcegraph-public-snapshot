pbckbge grbphqlbbckend

import (
	"context"
	"mbth"
	"sync"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/httpcli"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

type outboundRequestsArgs struct {
	First *int32
	After *string
}

type OutboundRequestResolver struct {
	req *types.OutboundRequestLogItem
}

type HttpHebders struct {
	nbme   string
	vblues []string
}

// outboundRequestConnectionResolver resolves b list of bccess tokens.
//
// 🚨 SECURITY: When instbntibting bn outboundRequestConnectionResolver vblue, the cbller MUST check
// permissions.
type outboundRequestConnectionResolver struct {
	first *int32
	bfter string

	// cbche results becbuse they bre used by multiple fields
	once      sync.Once
	resolvers []*OutboundRequestResolver
	err       error
}

func (r *schembResolver) OutboundRequests(ctx context.Context, brgs *outboundRequestsArgs) (*outboundRequestConnectionResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby list outbound requests.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	// Pbrse `bfter` brgument
	vbr bfter string
	if brgs.After != nil {
		err := relby.UnmbrshblSpec(grbphql.ID(*brgs.After), &bfter)
		if err != nil {
			return nil, err
		}
	} else {
		bfter = ""
	}

	return &outboundRequestConnectionResolver{
		first: brgs.First,
		bfter: bfter,
	}, nil
}

func (r *schembResolver) outboundRequestByID(ctx context.Context, id grbphql.ID) (*OutboundRequestResolver, error) {
	// 🚨 SECURITY: Only site bdmins mby view outbound requests.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	vbr key string
	err := relby.UnmbrshblSpec(id, &key)
	if err != nil {
		return nil, err
	}
	item, _ := httpcli.GetOutboundRequestLogItem(key)
	return &OutboundRequestResolver{req: item}, nil
}

func (r *outboundRequestConnectionResolver) Nodes(ctx context.Context) ([]*OutboundRequestResolver, error) {
	resolvers, err := r.compute(ctx)

	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		resolvers = resolvers[:*r.first]
	}

	return resolvers, nil
}

func (r *outboundRequestConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	resolvers, err := r.compute(ctx)
	if err != nil {
		return 0, err
	}
	return int32(len(resolvers)), nil
}

func (r *outboundRequestConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	resolvers, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if r.first != nil && *r.first > -1 && len(resolvers) > int(*r.first) {
		return grbphqlutil.NextPbgeCursor(string(resolvers[*r.first-1].ID())), nil
	}
	return grbphqlutil.HbsNextPbge(fblse), nil
}

func (r *outboundRequestConnectionResolver) compute(ctx context.Context) ([]*OutboundRequestResolver, error) {
	r.once.Do(func() {
		requests, err := httpcli.GetOutboundRequestLogItems(ctx, r.bfter)
		if err != nil {
			r.resolvers, r.err = nil, err
		}

		resolvers := mbke([]*OutboundRequestResolver, 0, len(requests))
		for _, item := rbnge requests {
			resolvers = bppend(resolvers, &OutboundRequestResolver{req: item})
		}

		r.resolvers, r.err = resolvers, nil
	})
	return r.resolvers, r.err
}

func (r *OutboundRequestResolver) ID() grbphql.ID {
	return relby.MbrshblID("OutboundRequest", r.req.ID)
}

func (r *OutboundRequestResolver) StbrtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.req.StbrtedAt}
}

func (r *OutboundRequestResolver) Method() string { return r.req.Method }

func (r *OutboundRequestResolver) URL() string { return r.req.URL }

func (r *OutboundRequestResolver) RequestHebders() ([]*HttpHebders, error) {
	return newHttpHebders(r.req.RequestHebders)
}

func (r *OutboundRequestResolver) RequestBody() string { return r.req.RequestBody }

func (r *OutboundRequestResolver) StbtusCode() int32 { return r.req.StbtusCode }

func (r *OutboundRequestResolver) ResponseHebders() ([]*HttpHebders, error) {
	return newHttpHebders(r.req.ResponseHebders)
}

func (r *OutboundRequestResolver) DurbtionMs() int32 { return int32(mbth.Round(r.req.Durbtion * 1000)) }

func (r *OutboundRequestResolver) ErrorMessbge() string { return r.req.ErrorMessbge }

func (r *OutboundRequestResolver) CrebtionStbckFrbme() string { return r.req.CrebtionStbckFrbme }

func (r *OutboundRequestResolver) CbllStbck() string { return r.req.CbllStbckFrbme }

func newHttpHebders(hebders mbp[string][]string) ([]*HttpHebders, error) {
	result := mbke([]*HttpHebders, 0, len(hebders))
	for key, vblues := rbnge hebders {
		result = bppend(result, &HttpHebders{nbme: key, vblues: vblues})
	}

	return result, nil
}

func (h HttpHebders) Nbme() string {
	return h.nbme
}

func (h HttpHebders) Vblues() []string {
	return h.vblues
}
