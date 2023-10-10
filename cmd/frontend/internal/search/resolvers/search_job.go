package resolvers

import (
	"context"
	"fmt"
	"net/url"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/service"
	"github.com/sourcegraph/sourcegraph/internal/search/exhaustive/types"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

const searchJobIDKind = "SearchJob"

func UnmarshalSearchJobID(id graphql.ID) (int64, error) {
	var v int64
	err := relay.UnmarshalSpec(id, &v)
	return v, err
}

var _ graphqlbackend.SearchJobResolver = &searchJobResolver{}

func newSearchJobResolver(db database.DB, svc *service.Service, job *types.ExhaustiveSearchJob) *searchJobResolver {
	return &searchJobResolver{Job: job, db: db, svc: svc}
}

// You should call newSearchJobResolver to construct an instance.
type searchJobResolver struct {
	Job *types.ExhaustiveSearchJob
	db  database.DB
	svc *service.Service
}

func (r *searchJobResolver) ID() graphql.ID {
	return relay.MarshalID(searchJobIDKind, r.Job.ID)
}

func (r *searchJobResolver) Query() string {
	return r.Job.Query
}

func (r *searchJobResolver) State(ctx context.Context) string {
	return r.Job.AggState.ToGraphQL()
}

func (r *searchJobResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := r.db.Users().GetByID(ctx, r.Job.InitiatorID)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewUserResolver(ctx, r.db, user), nil
}

func (r *searchJobResolver) CreatedAt() gqlutil.DateTime {
	return *gqlutil.FromTime(r.Job.CreatedAt)
}

func (r *searchJobResolver) StartedAt(ctx context.Context) *gqlutil.DateTime {
	return gqlutil.FromTime(r.Job.StartedAt)
}

func (r *searchJobResolver) FinishedAt(ctx context.Context) *gqlutil.DateTime {
	return gqlutil.FromTime(r.Job.FinishedAt)
}

func (r *searchJobResolver) URL(ctx context.Context) (*string, error) {
	if r.Job.State == types.JobStateCompleted {
		exportPath, err := url.JoinPath(conf.Get().ExternalURL, fmt.Sprintf("/.api/search/export/%d.csv", r.Job.ID))
		if err != nil {
			return nil, err
		}
		return pointers.Ptr(exportPath), nil
	}
	return nil, nil
}

func (r *searchJobResolver) LogURL(ctx context.Context) (*string, error) {
	if r.Job.State == types.JobStateCompleted {
		exportPath, err := url.JoinPath(conf.Get().ExternalURL, fmt.Sprintf("/.api/search/export/%d.log", r.Job.ID))
		if err != nil {
			return nil, err
		}
		return pointers.Ptr(exportPath), nil
	}
	return nil, nil
}

func (r *searchJobResolver) RepoStats(ctx context.Context) (graphqlbackend.SearchJobStatsResolver, error) {
	repoRevStats, err := r.svc.GetAggregateRepoRevState(ctx, r.Job.ID)
	if err != nil {
		return nil, err
	}
	return &searchJobStatsResolver{repoRevStats}, nil
}
