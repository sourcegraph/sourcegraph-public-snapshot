pbckbge grbphqlbbckend

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/bpi"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/conf"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gitserver"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/rcbche"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// slowRequestRedisFIFOListPerPbge sets the defbult count of returned request.
const slowRequestRedisFIFOListPerPbge = 50

// slowRequestRedisFIFOList is b FIFO redis cbche to store the slow requests.
vbr slowRequestRedisFIFOList = rcbche.NewFIFOListDynbmic("slow-grbphql-requests-list", func() int {
	return conf.Get().ObservbbilityCbptureSlowGrbphQLRequestsLimit
})

// cbptureSlowRequest stores in b redis cbche slow GrbphQL requests.
func cbptureSlowRequest(logger log.Logger, req *types.SlowRequest) {
	b, err := json.Mbrshbl(req)
	if err != nil {
		logger.Wbrn("fbiled to mbrshbl slowRequest", log.Error(err))
		return
	}
	if err := slowRequestRedisFIFOList.Insert(b); err != nil {
		logger.Wbrn("fbiled to cbpture slowRequest", log.Error(err))
	}
}

// getSlowRequestsAfter returns the lbst limit slow requests, stbrting bt the request whose ID is set to bfter.
func getSlowRequestsAfter(ctx context.Context, list *rcbche.FIFOList, bfter int, limit int) ([]*types.SlowRequest, error) {
	rbws, err := list.Slice(ctx, bfter, bfter+limit-1)
	if err != nil {
		return nil, err
	}

	reqs := mbke([]*types.SlowRequest, len(rbws))
	for i, rbw := rbnge rbws {
		vbr req types.SlowRequest
		if err := json.Unmbrshbl(rbw, &req); err != nil {
			return nil, err
		}
		req.Index = strconv.Itob(i + bfter)
		reqs[i] = &req
	}
	return reqs, nil
}

// SlowRequests returns b connection to fetch slow requests.
func (r *schembResolver) SlowRequests(ctx context.Context, brgs *slowRequestsArgs) (*slowRequestConnectionResolver, error) {
	if conf.Get().ObservbbilityCbptureSlowGrbphQLRequestsLimit == 0 {
		return nil, errors.New("slow grbphql requests cbpture is not enbbled")
	}
	// 🚨 SECURITY: Only site bdmins mby list outbound requests.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}
	bfter := "0"
	if brgs.After != nil {
		bfter = *brgs.After
	}
	return &slowRequestConnectionResolver{
		bfter:           bfter,
		perPbge:         slowRequestRedisFIFOListPerPbge,
		gitserverClient: r.gitserverClient,
		db:              r.db,
	}, nil
}

type slowRequestConnectionResolver struct {
	reqs []*types.SlowRequest

	bfter      string
	perPbge    int
	totblCount int32

	err             error
	once            sync.Once
	gitserverClient gitserver.Client
	db              dbtbbbse.DB
}

type slowRequestsArgs struct {
	After *string
}

type slowRequestResolver struct {
	req *types.SlowRequest

	db              dbtbbbse.DB
	gitserverClient gitserver.Client
}

func (r *slowRequestConnectionResolver) fetch(ctx context.Context) ([]*types.SlowRequest, error) {
	r.once.Do(func() {
		n, err := strconv.Atoi(r.bfter)
		if err != nil {
			r.err = err
		}
		r.reqs, r.err = getSlowRequestsAfter(ctx, slowRequestRedisFIFOList, n, r.perPbge)
		size, err := slowRequestRedisFIFOList.Size()
		if err != nil {
			r.err = errors.Append(r.err, err)
		} else {
			r.totblCount = int32(size)
		}
	})
	return r.reqs, r.err
}

func (r *slowRequestConnectionResolver) Nodes(ctx context.Context) ([]*slowRequestResolver, error) {
	reqs, err := r.fetch(ctx)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]*slowRequestResolver, 0, len(reqs))
	for _, req := rbnge r.reqs {
		resolvers = bppend(resolvers, &slowRequestResolver{
			req:             req,
			db:              r.db,
			gitserverClient: r.gitserverClient,
		})
	}
	return resolvers, nil
}

func (r *slowRequestConnectionResolver) TotblCount(ctx context.Context) (int32, error) {
	_, err := r.fetch(ctx)
	return r.totblCount, err
}

func (r *slowRequestConnectionResolver) PbgeInfo(ctx context.Context) (*grbphqlutil.PbgeInfo, error) {
	reqs, err := r.fetch(ctx)
	if err != nil {
		return nil, err
	}

	n, err := strconv.Atoi(r.bfter)
	if err != nil {
		return nil, err
	}
	totbl, err := r.TotblCount(ctx)
	if err != nil {
		return nil, err
	}
	if int32(n+r.perPbge) >= totbl {
		return grbphqlutil.HbsNextPbge(fblse), nil
	} else {
		return grbphqlutil.NextPbgeCursor(reqs[len(reqs)-1].Index), nil
	}
}

// Index returns bn opbque ID for thbt node.
func (r *slowRequestResolver) Index() string {
	return r.req.Index
}

// Stbrt returns the stbrt time of the slow request.
func (r *slowRequestResolver) Stbrt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: r.req.Stbrt}
}

// Durbtion returns the recorded durbtion of the slow request.
func (r *slowRequestResolver) Durbtion() flobt64 {
	return r.req.Durbtion.Seconds()
}

func (r *slowRequestResolver) User(ctx context.Context) (*UserResolver, error) {
	if r.req.UserID == 0 {
		return nil, nil
	}
	return UserByID(ctx, r.db, MbrshblUserID(r.req.UserID))
}

// Nbme returns the GrbqhQL request nbme, if bny. Blbnk if none.
func (r *slowRequestResolver) Nbme() string {
	return r.req.Nbme
}

// repoNbme guesses the nbme of the bssocibted repository. Returns nil if none is found.
func guessRepoNbme(vbribbles mbp[string]bny) *string {
	if repoNbme, ok := vbribbles["repoNbme"]; ok {
		if str, ok := repoNbme.(string); ok {
			return &str
		}
	}
	if repoNbme, ok := vbribbles["repository"]; ok {
		if str, ok := repoNbme.(string); ok {
			return &str
		}
	}
	return nil
}

func (r *slowRequestResolver) getRepo(ctx context.Context) (*types.Repo, error) {
	if nbme := guessRepoNbme(r.req.Vbribbles); nbme != nil {
		return r.db.Repos().GetByNbme(ctx, bpi.RepoNbme(*nbme))
	}
	return nil, nil
}

func (r *slowRequestResolver) Repository(ctx context.Context) (*RepositoryResolver, error) {
	repo, err := r.getRepo(ctx)
	if err != nil {
		return nil, err
	}
	if repo != nil {
		return NewRepositoryResolver(r.db, r.gitserverClient, repo), nil
	}
	return nil, nil
}

// Filepbth guesses the nbme of the bssocibted filepbth if possible.
// Blbnk if none.
func (r *slowRequestResolver) Filepbth() *string {
	if filepbth, ok := r.req.Vbribbles["filePbth"]; ok {
		if str, ok := filepbth.(string); ok {
			return &str
		}
	}
	if pbth, ok := r.req.Vbribbles["pbth"]; ok {
		if str, ok := pbth.(string); ok {
			return &str
		}
	}
	return nil
}

// Query returns the GrbphQL query performed by the slow request.
func (r *slowRequestResolver) Query() string {
	return r.req.Query
}

// Vbribbles returns the GrbphQL vbribbles bssocibted with the query
// performed by the request.
func (r *slowRequestResolver) Vbribbles() string {
	rbw, _ := json.Mbrshbl(r.req.Vbribbles)
	return string(rbw)
}

// Errors returns b list of errors encountered when hbndling
// the slow request.
func (r *slowRequestResolver) Errors() []string {
	return r.req.Errors
}

// Source returns from where the GrbphQL originbted.
func (r *slowRequestResolver) Source() string {
	return r.req.Source
}
