package campaigns

import (
	"context"
	"path"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threads"
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
	dbID, err := graphqlbackend.UnmarshalCampaignID(id)
	if err != nil {
		return nil, err
	}
	return campaignByDBID(ctx, dbID)
}

var MockCampaignByID func(graphql.ID) (graphqlbackend.Campaign, error)

func (GraphQLResolver) CampaignByID(ctx context.Context, id graphql.ID) (graphqlbackend.Campaign, error) {
	if MockCampaignByID != nil {
		return MockCampaignByID(id)
	}
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
		PartialComment: comments.GraphQLResolver{}.LazyCommentByID(graphqlbackend.MarshalCampaignID(v.ID)),
	}
}

func (v *gqlCampaign) ID() graphql.ID {
	return graphqlbackend.MarshalCampaignID(v.db.ID)
}

func (v *gqlCampaign) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return graphqlbackend.NamespaceByDBID(ctx, v.db.NamespaceUserID, v.db.NamespaceOrgID)
}

func (v *gqlCampaign) Name() string { return v.db.Name }

func (v *gqlCampaign) IsPreview() bool { return v.db.IsPreview }

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

func (v *gqlCampaign) Rules(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.RuleConnection, error) {
	return graphqlbackend.RulesInRuleContainer(ctx, v.ID(), arg)
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

func (v *gqlCampaign) Threads(ctx context.Context, arg *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadConnection, error) {
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
	return threads.ThreadsByIDs(threadIDs, arg), nil
}

func (v *gqlCampaign) getThreads(ctx context.Context) ([]graphqlbackend.Thread, error) {
	connection, err := v.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	return connection.Nodes(ctx)
}

func (v *gqlCampaign) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	threads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	return campaignRepositories(ctx, threads)
}

func (v *gqlCampaign) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	threads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	return campaignCommits(ctx, threads)
}

func (v *gqlCampaign) RepositoryComparisons(ctx context.Context) ([]*graphqlbackend.RepositoryComparisonResolver, error) {
	threads, err := v.getThreads(ctx)
	if err != nil {
		return nil, err
	}
	return campaignRepositoryComparisons(ctx, threads)
}

func (v *gqlCampaign) Diagnostics(ctx context.Context, arg *graphqlbackend.ThreadDiagnosticConnectionArgs) (graphqlbackend.ThreadDiagnosticConnection, error) {
	campaignID := v.ID()
	arg.Campaign = &campaignID
	return graphqlbackend.ThreadDiagnostics.ThreadDiagnostics(ctx, arg)
}
