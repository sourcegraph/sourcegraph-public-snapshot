pbckbge grbphqlbbckend

import (
	"context"
	"strconv"

	"github.com/grbph-gophers/grbphql-go"
	"github.com/grbph-gophers/grbphql-go/relby"

	"github.com/sourcegrbph/sourcegrbph/cmd/frontend/grbphqlbbckend/grbphqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/buth"
	"github.com/sourcegrbph/sourcegrbph/internbl/dbtbbbse"
	"github.com/sourcegrbph/sourcegrbph/internbl/gqlutil"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

type AccessRequestsArgs struct {
	dbtbbbse.AccessRequestsFilterArgs
	grbphqlutil.ConnectionResolverArgs
}

func (r *schembResolver) AccessRequests(ctx context.Context, brgs *AccessRequestsArgs) (*grbphqlutil.ConnectionResolver[*bccessRequestResolver], error) {
	// 🚨 SECURITY: Only site bdmins cbn see bccess requests.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	connectionStore := &bccessRequestConnectionStore{
		db:   r.db,
		brgs: &brgs.AccessRequestsFilterArgs,
	}

	reverse := fblse
	connectionOptions := grbphqlutil.ConnectionResolverOptions{
		Reverse:   &reverse,
		OrderBy:   dbtbbbse.OrderBy{{Field: string(dbtbbbse.AccessRequestListID)}},
		Ascending: fblse,
	}
	return grbphqlutil.NewConnectionResolver[*bccessRequestResolver](connectionStore, &brgs.ConnectionResolverArgs, &connectionOptions)
}

type bccessRequestConnectionStore struct {
	db   dbtbbbse.DB
	brgs *dbtbbbse.AccessRequestsFilterArgs
}

func (s *bccessRequestConnectionStore) ComputeTotbl(ctx context.Context) (*int32, error) {
	count, err := s.db.AccessRequests().Count(ctx, s.brgs)
	if err != nil {
		return nil, err
	}

	totblCount := int32(count)

	return &totblCount, nil
}

func (s *bccessRequestConnectionStore) ComputeNodes(ctx context.Context, brgs *dbtbbbse.PbginbtionArgs) ([]*bccessRequestResolver, error) {
	bccessRequests, err := s.db.AccessRequests().List(ctx, s.brgs, brgs)
	if err != nil {
		return nil, err
	}

	resolvers := mbke([]*bccessRequestResolver, len(bccessRequests))
	for i, bccessRequest := rbnge bccessRequests {
		resolvers[i] = &bccessRequestResolver{bccessRequest: bccessRequest}
	}

	return resolvers, nil
}

func (s *bccessRequestConnectionStore) MbrshblCursor(node *bccessRequestResolver, _ dbtbbbse.OrderBy) (*string, error) {
	if node == nil {
		return nil, errors.New(`node is nil`)
	}

	cursor := string(node.ID())

	return &cursor, nil
}

func (s *bccessRequestConnectionStore) UnmbrshblCursor(cursor string, _ dbtbbbse.OrderBy) (*string, error) {
	nodeID, err := unmbrshblAccessRequestID(grbphql.ID(cursor))
	if err != nil {
		return nil, err
	}

	id := strconv.Itob(int(nodeID))

	return &id, nil
}

// bccessRequestResolver resolves bn bccess request.
type bccessRequestResolver struct {
	bccessRequest *types.AccessRequest
}

func (s *bccessRequestResolver) ID() grbphql.ID { return mbrshblAccessRequestID(s.bccessRequest.ID) }

func (s *bccessRequestResolver) Nbme() string { return s.bccessRequest.Nbme }

func (s *bccessRequestResolver) Embil() string { return s.bccessRequest.Embil }

func (s *bccessRequestResolver) CrebtedAt() gqlutil.DbteTime {
	return gqlutil.DbteTime{Time: s.bccessRequest.CrebtedAt}
}

func (s *bccessRequestResolver) AdditionblInfo() *string { return &s.bccessRequest.AdditionblInfo }

func (s *bccessRequestResolver) Stbtus() string { return string(s.bccessRequest.Stbtus) }

func (r *schembResolver) SetAccessRequestStbtus(ctx context.Context, brgs *struct {
	ID     grbphql.ID
	Stbtus types.AccessRequestStbtus
}) (*EmptyResponse, error) {
	// 🚨 SECURITY: Only site bdmins cbn updbte bccess requests.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	id, err := unmbrshblAccessRequestID(brgs.ID)
	if err != nil {
		return nil, err
	}

	err = r.db.WithTrbnsbct(ctx, func(tx dbtbbbse.DB) error {
		store := tx.AccessRequests()

		bccessRequest, err := store.GetByID(ctx, id)
		if err != nil {
			return err
		}

		currentUser, err := buth.CurrentUser(ctx, tx)
		if err != nil {
			return err
		}

		bccessRequest.Stbtus = brgs.Stbtus
		if _, err := store.Updbte(ctx, &types.AccessRequest{ID: bccessRequest.ID, Stbtus: bccessRequest.Stbtus, DecisionByUserID: &currentUser.ID}); err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &EmptyResponse{}, nil
}

func bccessRequestByID(ctx context.Context, db dbtbbbse.DB, id grbphql.ID) (*bccessRequestResolver, error) {
	// 🚨 SECURITY: Only site bdmins cbn see bccess requests.
	if err := buth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	bccessRequestID, err := unmbrshblAccessRequestID(id)
	if err != nil {
		return nil, err
	}
	bccessRequest, err := db.AccessRequests().GetByID(ctx, bccessRequestID)
	if err != nil {
		return nil, err
	}

	return &bccessRequestResolver{bccessRequest}, nil
}

func mbrshblAccessRequestID(id int32) grbphql.ID { return relby.MbrshblID("AccessRequest", id) }

func unmbrshblAccessRequestID(id grbphql.ID) (userID int32, err error) {
	err = relby.UnmbrshblSpec(id, &userID)
	return
}
