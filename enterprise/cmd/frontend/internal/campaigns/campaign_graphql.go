package campaigns

import (
	"context"
	"path"
	"sort"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
	"github.com/sourcegraph/sourcegraph/pkg/api"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// gqlCampaign implements the GraphQL type Campaign.
type gqlCampaign struct {
	db *dbCampaign
	graphqlbackend.PartialComment
}

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
	return newGQLCampaign(v), nil
}

func (GraphQLResolver) CampaignByDBID(ctx context.Context, id int64) (graphqlbackend.Campaign, error) {
	return campaignByDBID(ctx, id)
}

func newGQLCampaign(v *dbCampaign) *gqlCampaign {
	return &gqlCampaign{
		db:             v,
		PartialComment: comments.GraphQLResolver{}.LazyCommentByID(marshalCampaignID(v.ID)),
	}
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

func (v *gqlCampaign) IsPreview() bool { return v.db.IsPreview }

func (v *gqlCampaign) Rules() string {
	if v.db.Rules != "" {
		return v.db.Rules
	}
	return "[]"
}

func (v *gqlCampaign) ViewerCanUpdate(ctx context.Context) (bool, error) {
	return commentobjectdb.ViewerCanUpdate(ctx, v.ID())
}

func (v *gqlCampaign) ViewerCanComment(ctx context.Context) (bool, error) {
	return commentobjectdb.ViewerCanComment(ctx)
}

func (v *gqlCampaign) ViewerCannotCommentReasons(ctx context.Context) ([]graphqlbackend.CannotCommentReason, error) {
	return commentobjectdb.ViewerCannotCommentReasons(ctx)
}

func (v *gqlCampaign) Comments(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.CommentConnection, error) {
	return graphqlbackend.CommentsForObject(ctx, v.ID(), arg)
}

func (v *gqlCampaign) URL(ctx context.Context) (string, error) {
	namespace, err := v.Namespace(ctx)
	if err != nil {
		return "", err
	}

	var preview string
	if v.db.IsPreview {
		preview = "preview"
	}

	return path.Join(namespace.URL(), "campaigns", preview, string(v.ID())), nil
	//
	// TODO!(sqs): use global url?
	// return path.Join("/campaigns", string(v.ID())), nil
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
	return threads.ThreadsByIDs(threadIDs), nil
}

func (v *gqlCampaign) getThreads(ctx context.Context) ([]graphqlbackend.Thread, error) {
	connection, err := v.Threads(ctx, &graphqlutil.ConnectionArgs{})
	if err != nil {
		return nil, err
	}
	return connection.Nodes(ctx)
}

func (v *gqlCampaign) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	threadNodes, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}

	byRepositoryDBID := map[api.RepoID]*graphqlbackend.RepositoryResolver{}
	for _, thread := range threadNodes {
		repo, err := thread.Repository(ctx)
		if err != nil {
			return nil, err
		}
		key := repo.DBID()
		if _, seen := byRepositoryDBID[key]; !seen {
			byRepositoryDBID[key] = repo
		}
	}

	repos := make([]*graphqlbackend.RepositoryResolver, 0, len(byRepositoryDBID))
	for _, repo := range byRepositoryDBID {
		repos = append(repos, repo)
	}
	sort.Slice(repos, func(i, j int) bool {
		return repos[i].DBID() < repos[j].DBID()
	})
	return repos, nil
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
	threads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}

	rcs := make([]*graphqlbackend.RepositoryComparisonResolver, len(threads))
	for i, thread := range threads {
		rc, err := thread.RepositoryComparison(ctx)
		if err != nil {
			return nil, err
		}
		rcs[i] = rc
	}
	return rcs, nil
}
