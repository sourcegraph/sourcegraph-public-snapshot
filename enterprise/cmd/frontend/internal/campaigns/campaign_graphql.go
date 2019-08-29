package campaigns

import (
	"context"
	"path"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
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

func (v *gqlCampaign) DBID() int64 { return v.db.ID }

func (v *gqlCampaign) Namespace(ctx context.Context) (*graphqlbackend.NamespaceResolver, error) {
	return graphqlbackend.NamespaceByDBID(ctx, v.db.NamespaceUserID, v.db.NamespaceOrgID)
}

func (v *gqlCampaign) Name() string { return v.db.Name }

func (v *gqlCampaign) Template() *graphqlbackend.CampaignTemplateInstance {
	if v.db.TemplateID != nil {
		var context graphqlbackend.JSONC
		if v.db.TemplateContext != nil {
			context = graphqlbackend.JSONC(*v.db.TemplateContext)
		} else {
			context = "{}"
		}
		return &graphqlbackend.CampaignTemplateInstance{
			Template_: *v.db.TemplateID,
			Context_:  context,
		}
	}
	return nil
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

func (v *gqlCampaign) Rules(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.RuleConnection, error) {
	return graphqlbackend.RulesInRuleContainer(ctx, v.ID(), arg)
}

func (v *gqlCampaign) IsDraft() bool { return v.db.IsDraft }

func (v *gqlCampaign) StartDate() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(v.db.StartDate)
}

func (v *gqlCampaign) DueDate() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(v.db.DueDate)
}

func (v *gqlCampaign) URL(ctx context.Context) (string, error) {
	namespace, err := v.Namespace(ctx)
	if err != nil {
		return "", err
	}
	return path.Join(namespace.URL(), "campaigns", string(v.ID())), nil
	//
	// TODO!(sqs): use global url?
	// return path.Join("/campaigns", string(v.ID())), nil
}

func (v *gqlCampaign) Threads(ctx context.Context, arg *graphqlbackend.ThreadConnectionArgs) (graphqlbackend.ThreadOrThreadPreviewConnection, error) {
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
	c, err := threads.ThreadsByIDs(ctx, threadIDs, arg)
	if err != nil {
		return nil, err
	}
	return toThreadOrThreadPreviewConnection{c}, nil
}

type toThreadOrThreadPreviewConnection struct {
	graphqlbackend.ThreadConnection
}

func (c toThreadOrThreadPreviewConnection) Nodes(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	nodes, err := c.ThreadConnection.Nodes(ctx)
	if err != nil {
		return nil, err
	}
	return threads.ToThreadOrThreadPreviews(nodes, nil), nil
}

func (v *gqlCampaign) getThreads(ctx context.Context) ([]graphqlbackend.ToThreadOrThreadPreview, error) {
	connection, err := v.Threads(ctx, &graphqlbackend.ThreadConnectionArgs{})
	if err != nil {
		return nil, err
	}
	return connection.Nodes(ctx)
}

func (v *gqlCampaign) Repositories(ctx context.Context) ([]*graphqlbackend.RepositoryResolver, error) {
	return campaignRepositories(ctx, v)
}

func (v *gqlCampaign) Commits(ctx context.Context) ([]*graphqlbackend.GitCommitResolver, error) {
	return campaignCommits(ctx, v)
}

func (v *gqlCampaign) RepositoryComparisons(ctx context.Context) ([]graphqlbackend.RepositoryComparison, error) {
	return campaignRepositoryComparisons(ctx, v)
}

func (v *gqlCampaign) Diagnostics(ctx context.Context, arg *graphqlbackend.ThreadDiagnosticConnectionArgs) (graphqlbackend.ThreadDiagnosticConnection, error) {
	campaignID := v.ID()
	arg.Campaign = &campaignID
	return graphqlbackend.ThreadDiagnostics.ThreadDiagnostics(ctx, arg)
}

func (v *gqlCampaign) BurndownChart(ctx context.Context) (graphqlbackend.CampaignBurndownChart, error) {
	return campaignBurndownChart(ctx, v)
}

func (v *gqlCampaign) getEvents(ctx context.Context, beforeDate time.Time, eventTypes []events.Type) ([]graphqlbackend.ToEvent, error) {
	eventTypes2 := make([]string, len(eventTypes))
	for i, t := range eventTypes {
		eventTypes2[i] = string(t)
	}
	ec, err := events.GetEventConnection(ctx,
		&graphqlbackend.EventConnectionCommonArgs{
			BeforeDate: &graphqlbackend.DateTime{beforeDate},
			Types:      &eventTypes2,
		},
		events.Objects{Campaign: v.db.ID},
	)
	if err != nil {
		return nil, err
	}
	return ec.Nodes(ctx)
}

func (v *gqlCampaign) TimelineItems(ctx context.Context, arg *graphqlbackend.EventConnectionCommonArgs) (graphqlbackend.EventConnection, error) {
	return events.GetEventConnection(ctx,
		arg,
		events.Objects{Campaign: v.db.ID},
	)
}

func (v *gqlCampaign) Participants(ctx context.Context, arg *graphqlbackend.ParticipantConnectionArgs) (graphqlbackend.ParticipantConnection, error) {
	return campaignParticipants(ctx, v)
}
