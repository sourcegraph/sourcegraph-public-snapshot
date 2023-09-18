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

type searchJobResolver struct {
	Job *types.ExhaustiveSearchJob
	db  database.DB
}

func (r searchJobResolver) ID() graphql.ID {
	return relay.MarshalID(searchJobIDKind, r.Job.ID)
}

func (r searchJobResolver) Query() string {
	return r.Job.Query
}

func (r searchJobResolver) State() string {
	return r.Job.State.ToGraphQL()
}

func (r searchJobResolver) Creator(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	user, err := r.db.Users().GetByID(ctx, r.Job.InitiatorID)
	if err != nil {
		return nil, err
	}
	return graphqlbackend.NewUserResolver(ctx, r.db, user), nil
}

func (r searchJobResolver) CreatedAt() gqlutil.DateTime {
	return *gqlutil.FromTime(r.Job.CreatedAt)
}

func (r searchJobResolver) StartedAt(ctx context.Context) *gqlutil.DateTime {
	return gqlutil.FromTime(r.Job.StartedAt)
}

func (r searchJobResolver) FinishedAt(ctx context.Context) *gqlutil.DateTime {
	return gqlutil.FromTime(r.Job.FinishedAt)
}

func (r searchJobResolver) URL(ctx context.Context) (*string, error) {
	if r.Job.State == types.JobStateCompleted {
		exportPath, err := url.JoinPath(conf.Get().ExternalURL, fmt.Sprintf("/.api/search/export/%d", r.Job.ID))
		if err != nil {
			return nil, err
		}
		return pointers.Ptr(exportPath), nil
	}
	return nil, nil
}

func (r searchJobResolver) RepoStats(ctx context.Context) (graphqlbackend.SearchJobStatsResolver, error) {
	// TODO: This needs to be implemented properly, this is fake data
	stats := &searchJobStatsResolver{
		total:      99,
		completed:  80,
		failed:     10,
		inProgress: 9,
	}

	return stats, nil
}
