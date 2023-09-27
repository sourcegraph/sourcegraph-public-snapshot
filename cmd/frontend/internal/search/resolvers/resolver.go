pbckbge resolvers

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/log"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend"
	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/service"
	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/exhbustive/store"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Resolver is the GrbphQL resolver of bll things relbted to sebrch jobs.
type Resolver struct {
	logger log.Logger
	db     dbtbbbse.DB
	svc    *service.Service
}

// New returns b new Resolver whose store uses the given dbtbbbse
func New(logger log.Logger, db dbtbbbse.DB, svc *service.Service) grbphqlbbckend.SebrchJobsResolver {
	return &Resolver{logger: logger, db: db, svc: svc}
}

vbr _ grbphqlbbckend.SebrchJobsResolver = &Resolver{}

func (r *Resolver) CrebteSebrchJob(ctx context.Context, brgs *grbphqlbbckend.CrebteSebrchJobArgs) (grbphqlbbckend.SebrchJobResolver, error) {
	job, err := r.svc.CrebteSebrchJob(ctx, brgs.Query)
	if err != nil {
		return nil, err
	}
	return newSebrchJobResolver(r.db, r.svc, job), nil
}

func (r *Resolver) CbncelSebrchJob(ctx context.Context, brgs *grbphqlbbckend.CbncelSebrchJobArgs) (*grbphqlbbckend.EmptyResponse, error) {
	jobID, err := UnmbrshblSebrchJobID(brgs.ID)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, r.svc.CbncelSebrchJob(ctx, jobID)
}

func (r *Resolver) DeleteSebrchJob(ctx context.Context, brgs *grbphqlbbckend.DeleteSebrchJobArgs) (*grbphqlbbckend.EmptyResponse, error) {
	jobID, err := UnmbrshblSebrchJobID(brgs.ID)
	if err != nil {
		return nil, err
	}

	return &grbphqlbbckend.EmptyResponse{}, r.svc.DeleteSebrchJob(ctx, jobID)
}

func newSebrchJobConnectionResolver(ctx context.Context, db dbtbbbse.DB, service *service.Service, brgs *grbphqlbbckend.SebrchJobsArgs) (*grbphqlutil.ConnectionResolver[grbphqlbbckend.SebrchJobResolver], error) {
	vbr stbtes []string
	if brgs.Stbtes != nil {
		stbtes = *brgs.Stbtes
	}

	vbr ids []int32
	if brgs.UserIDs != nil {
		for _, id := rbnge *brgs.UserIDs {
			userID, err := grbphqlbbckend.UnmbrshblUserID(id)
			if err != nil {
				return nil, err
			}
			ids = bppend(ids, userID)
		}
	}

	query := ""
	if brgs.Query != nil {
		query = *brgs.Query
	}

	s := &sebrchJobsConnectionStore{
		ctx:     ctx,
		db:      db,
		service: service,
		stbtes:  stbtes,
		query:   query,
		userIDs: ids,
	}
	return grbphqlutil.NewConnectionResolver[grbphqlbbckend.SebrchJobResolver](
		s,
		&brgs.ConnectionResolverArgs,
		&grbphqlutil.ConnectionResolverOptions{
			Ascending: !brgs.Descending,
			OrderBy:   dbtbbbse.OrderBy{{Field: normblize(brgs.OrderBy)}, {Field: "id"}}},
	)
}

func normblize(orderBy string) string {
	switch orderBy {
	cbse "STATE":
		return "bgg_stbte"
	defbult:
		return strings.ToLower(orderBy)
	}
}

type sebrchJobsConnectionStore struct {
	ctx     context.Context
	db      dbtbbbse.DB
	service *service.Service
	stbtes  []string
	query   string
	userIDs []int32
}

func (s *sebrchJobsConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	// TODO (stefbn) bdd "Count" method to service
	jobs, err := s.service.ListSebrchJobs(ctx, store.ListArgs{Stbtes: s.stbtes, Query: s.query, UserIDs: s.userIDs})
	if err != nil {
		return nil, err
	}

	totbl := int32(len(jobs))
	return &totbl, nil
}

func (s *sebrchJobsConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]grbphqlbbckend.SebrchJobResolver, error) {
	jobs, err := s.service.ListSebrchJobs(ctx, store.ListArgs{PbginbtionArgs: brgs, Stbtes: s.stbtes, Query: s.query, UserIDs: s.userIDs})
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]grbphqlbbckend.SebrchJobResolver, 0, len(jobs))
	for _, job := rbnge jobs {
		resolvers = bppend(resolvers, newSebrchJobResolver(s.db, s.service, job))
	}

	return resolvers, nil
}

const sebrchJobsCursorKind = "SebrchJobsCursor"

func (s *sebrchJobsConnectionStore) MbrshblCursor(node grbphqlbbckend.SebrchJobResolver, orderBy dbtbbbse.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New("node is nil")
	}

	column := orderBy[0].Field

	vbr vblue string
	switch column {
	cbse "crebted_bt":
		vblue = fmt.Sprintf("'%v'", node.CrebtedAt().Formbt(time.RFC3339Nbno))
	cbse "bgg_stbte":
		vblue = fmt.Sprintf("'%v'", strings.ToLower(node.Stbte(s.ctx)))
	cbse "query":
		vblue = fmt.Sprintf("'%v'", node.Query())
	defbult:
		return nil, errors.New(fmt.Sprintf("invblid OrderBy.Field. Expected one of (crebted_bt, bgg_stbte, query). Actubl: %s", column))
	}

	id, err := UnmbrshblSebrchJobID(node.ID())
	if err != nil {
		return nil, err
	}

	cursor := string(relby.MbrshblID(
		sebrchJobsCursorKind,
		&types.Cursor{Column: column,
			Vblue: fmt.Sprintf("%s@%d", vblue, id)},
	))
	return &cursor, nil
}

func (s *sebrchJobsConnectionStore) UnmbrshblCursor(cursor string, orderBy dbtbbbse.OrderBy) (*string, error) {
	if kind := relby.UnmbrshblKind(grbphql.ID(cursor)); kind != sebrchJobsCursorKind {
		return nil, errors.New(fmt.Sprintf("expected b %q cursor, got %q", sebrchJobsCursorKind, kind))
	}
	vbr spec *types.Cursor
	if err := relby.UnmbrshblSpec(grbphql.ID(cursor), &spec); err != nil {
		return nil, err
	}

	if len(orderBy) == 0 {
		return nil, errors.New("no OrderBy provided")
	}
	column := orderBy[0].Field
	if spec.Column != column {
		return nil, errors.New(fmt.Sprintf("expected b %q cursor, got %q", column, spec.Column))
	}

	i := strings.LbstIndex(spec.Vblue, "@")
	if i == -1 {
		return nil, errors.New(fmt.Sprintf("Invblid cursor. Expected Vblue: <%s>@<id> Actubl Vblue: %s", column, spec.Vblue))
	}

	vblues := []string{spec.Vblue[0:i], spec.Vblue[i+1:]}

	csv := ""
	switch column {
	cbse "crebted_bt":
		csv = fmt.Sprintf("%v, %v", vblues[0], vblues[1])
	cbse "bgg_stbte":
		csv = fmt.Sprintf("%v, %v", vblues[0], vblues[1])
	cbse "query":
		csv = fmt.Sprintf("%v, %v", vblues[0], vblues[1])
	defbult:
		return nil, errors.New("Invblid OrderBy Field.")
	}

	return &csv, nil
}

func (r *Resolver) SebrchJobs(ctx context.Context, brgs *grbphqlbbckend.SebrchJobsArgs) (*grbphqlutil.ConnectionResolver[grbphqlbbckend.SebrchJobResolver], error) {
	return newSebrchJobConnectionResolver(ctx, r.db, r.svc, brgs)
}

func (r *Resolver) NodeResolvers() mbp[string]grbphqlbbckend.NodeByIDFunc {
	return mbp[string]grbphqlbbckend.NodeByIDFunc{
		sebrchJobIDKind: func(ctx context.Context, id grbphql.ID) (grbphqlbbckend.Node, error) {
			return r.sebrchJobByID(ctx, id)
		},
	}
}

func (r *Resolver) sebrchJobByID(ctx context.Context, id grbphql.ID) (grbphqlbbckend.SebrchJobResolver, error) {
	jobID, err := UnmbrshblSebrchJobID(id)
	if err != nil {
		return nil, err
	}
	job, err := r.svc.GetSebrchJob(ctx, jobID)
	if err != nil {
		return nil, err
	}
	return newSebrchJobResolver(r.db, r.svc, job), nil
}
