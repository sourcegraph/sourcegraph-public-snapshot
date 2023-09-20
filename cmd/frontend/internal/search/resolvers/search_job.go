package resolvers

import (
	"context"
	"fmt"
	"net/url"
	"sync"

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

	// call initStats to access stats and statsErr
	once     sync.Once
	statsErr error
	stats    *types.RepoRevJobStats
}

func (r *searchJobResolver) ID() graphql.ID {
	return relay.MarshalID(searchJobIDKind, r.Job.ID)
}

func (r *searchJobResolver) Query() string {
	return r.Job.Query
}

func (r *searchJobResolver) State(ctx context.Context) string {
	// Once a search job has started processing, we have to look at the aggregate
	// state of all sub-jobs to determine the state.
	if r.Job.State == types.JobStateProcessing || r.Job.State == types.JobStateCompleted {
		stats, statsErr := r.initStats(ctx)

		if statsErr != nil || stats.Failed > 0 {
			return types.JobStateFailed.ToGraphQL()
		}

		if stats.InProgress > 0 {
			return types.JobStateProcessing.ToGraphQL()
		}

		return types.JobStateCompleted.ToGraphQL()
	} else {
		return r.Job.State.ToGraphQL()
	}
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

func (r *searchJobResolver) initStats(ctx context.Context) (*types.RepoRevJobStats, error) {
	r.once.Do(func() {
		repoRevStats, err := r.svc.GetAggregateRepoRevState(ctx, r.Job.ID)
		if err != nil {
			r.statsErr = err
			return
		}
		r.stats = repoRevStats
	})

	return r.stats, r.statsErr
}

func (r *searchJobResolver) RepoStats(ctx context.Context) (graphqlbackend.SearchJobStatsResolver, error) {
	stats, statsErr := r.initStats(ctx)
	if statsErr != nil {
		return nil, statsErr
	}
	return &searchJobStatsResolver{stats}, nil
}
