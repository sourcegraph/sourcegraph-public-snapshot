package campaigns

import (
	"context"
	"encoding/json"
	"path"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlCampaign implements the GraphQL type Campaign.
type gqlCampaign struct{ db *dbCampaign }

// campaignByID looks up and returns the Campaign with the given GraphQL ID. If no such Campaign exists, it
// returns a non-nil error.
func campaignByID(ctx context.Context, id graphql.ID) (*gqlCampaign, error) {
	dbID, err := unmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}
	return campaignByDBID(ctx, dbID)
}

func (GraphQLResolver) CampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.Campaign, error) {
	return campaignByID(ctx, id)
}

// campaignByDBID looks up and returns the Campaign with the given database ID. If no such Campaign exists,
// it returns a non-nil error.
func campaignByDBID(ctx context.Context, dbID int64) (*gqlCampaign, error) {
	v, err := dbCampaigns{}.GetByID(ctx, dbID)
	if err != nil {
		return nil, err
	}
	return &gqlCampaign{db: v}, nil
}

func (v *gqlCampaign) ID() graphql.ID {
	return marshalCampaignID(v.db.ID)
}

func marshalCampaignID(id int64) graphql.ID {
	return relay.MarshalID("Campaign", id)
}

func unmarshalCampaignID(id graphql.ID) (dbID int64, err error) {
	err = relay.UnmarshalSpec(id, &dbID)
	return
}

func (v *gqlCampaign) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return graphqlbackend.NamespaceByDBID(ctx, v.db.NamespaceUserID, v.db.NamespaceOrgID)
}

func (v *gqlCampaign) Name() string { return v.db.Name }

func (v *gqlCampaign) Description() *string { return v.db.Description }

func (v *gqlCampaign) URL(ctx context.Context) (string, error) {
	namespace, err := v.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(namespace.URL(), "campaigns", string(v.ID())), nil
}

func (v *gqlCampaign) Threads(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.ThreadConnection, error) {
	opt := dbCampaignsThreadsListOptions{CampaignID: v.db.ID}
	arg.Set(&opt.LimitOffset)
	l, err := dbCampaignsThreads{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}

	threadIDs := make([]int64, len(l))
	for i, e := range l {
		threadIDs[i] = e.Thread
	}
	return threads.ThreadsByIDs(ctx, threadIDs, arg)
}

// TODO!(sqs)
type delta struct {
	Repository graphql.ID
	Base, Head string
}

func (v *gqlCampaign) getDeltas(ctx context.Context) ([]*delta, error) {
	threadConnection, err := v.Threads(ctx, &graphqlutil.ConnectionArgs{})
	if err != nil {
		return nil, err
	}
	threads, err := threadConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}

	var deltas []*delta
	for _, thread := range threads {
		var settings struct{ Deltas []*delta }
		if err := json.Unmarshal([]byte(thread.Settings()), &settings); err != nil {
			return nil, err
		}
		deltas = append(deltas, settings.Deltas...)
	}

	return deltas, nil
}

func (v *gqlCampaign) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	deltas, err := v.getDeltas(ctx)
	if err != nil {
		return nil, err
	}

	rs := make([]*graphqlbackend.RepositoryResolver, len(deltas))
	for i, delta := range deltas {
		var err error
		rs[i], err = graphqlbackend.RepositoryByID(ctx, delta.Repository)
		if err != nil {
			return nil, err
		}
	}
	return rs, nil
}

func (v *gqlCampaign) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	rcs, err := v.RepositoryComparisons(ctx)
	if err != nil {
		return nil, err
	}

	var allCommits []*graphqlbackend.GitCommitResolver
	for _, rc := range rcs {
		cc := rc.Commits(&graphqlutil.ConnectionArgs{})
		commits, err := cc.Nodes(ctx)
		if err != nil {
			return nil, err
		}
		allCommits = append(allCommits, commits...)
	}
	return allCommits, nil
}

func (v *gqlCampaign) RepositoryComparisons(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	deltas, err := v.getDeltas(ctx)
	if err != nil {
		return nil, err
	}

	rcs := make([]*graphqlbackend.RepositoryComparisonResolver, len(deltas))
	for i, delta := range deltas {
		repo, err := graphqlbackend.RepositoryByID(ctx, delta.Repository)
		if err != nil {
			return nil, err
		}
		rcs[i], err = graphqlbackend.NewRepositoryComparison(ctx, repo, &graphqlbackend.RepositoryComparisonInput{
			Base: &delta.Base,
			Head: &delta.Head,
		})
		if err != nil {
			return nil, err
		}
	}
	return rcs, nil
}
