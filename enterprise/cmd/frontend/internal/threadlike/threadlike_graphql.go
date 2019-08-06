package threadlike

import (
	"context"
	"errors"
	"path"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/events"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// GQLThreadlike implements common fields for the GraphQL thread, issue, and changeset types.
type GQLThreadlike struct {
	DB *internal.DBThread
	graphqlbackend.PartialComment
}

func (v *GQLThreadlike) ID() graphql.ID {
	var gqlType gqlType
	switch v.DB.Type {
	case internal.DBThreadTypeThread:
		gqlType = GQLTypeThread
	case internal.DBThreadTypeIssue:
		gqlType = GQLTypeIssue
	case internal.DBThreadTypeChangeset:
		gqlType = GQLTypeChangeset
	default:
		panic("invalid thread type: " + v.DB.Type)
	}
	return MarshalID(gqlType, v.DB.ID)
}

func (v *GQLThreadlike) Type() internal.DBThreadType { return v.DB.Type }

func (v *GQLThreadlike) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, v.DB.RepositoryID)
}

func (v *GQLThreadlike) Number() string { return strconv.FormatInt(v.DB.ID, 10) }

func (v *GQLThreadlike) DBID() int64 { return v.DB.ID }

func (v *GQLThreadlike) Title() string { return v.DB.Title }

func (v *GQLThreadlike) ViewerCanUpdate(ctx context.Context) (bool, error) {
	// TODO!(sqs): commented out below due to package import cycle etc
	return true, nil
	// return commentobjectdb.ViewerCanUpdate(ctx, v.ID())
}

func (v *GQLThreadlike) ViewerCanComment(ctx context.Context) (bool, error) {
	return commentobjectdb.ViewerCanComment(ctx)
}

func (v *GQLThreadlike) ViewerCannotCommentReasons(ctx context.Context) ([]graphqlbackend.CannotCommentReason, error) {
	return commentobjectdb.ViewerCannotCommentReasons(ctx)
}

func (v *GQLThreadlike) URL(ctx context.Context) (string, error) {
	repository, err := v.Repository(ctx)
	if err != nil {
		return "", err
	}

	var typeComponent string
	switch v.DB.Type {
	case internal.DBThreadTypeThread:
		typeComponent = "threads"
	case internal.DBThreadTypeIssue:
		typeComponent = "issues"
	case internal.DBThreadTypeChangeset:
		typeComponent = "changesets"
	default:
		return "", errors.New("invalid thread type")
	}

	return path.Join(repository.URL(), "-", typeComponent, v.Number()), nil
}

func (v *GQLThreadlike) Campaigns(ctx context.Context, arg *graphqlutil.ConnectionArgs) (graphqlbackend.CampaignConnection, error) {
	return graphqlbackend.CampaignsWithObject(ctx, v.ID(), arg)
}

func (v *GQLThreadlike) TimelineItems(ctx context.Context, arg *graphqlbackend.EventConnectionCommonArgs) (graphqlbackend.EventConnection, error) {
	return events.GetEventConnection(ctx,
		arg,
		events.Objects{Thread: v.DB.ID},
	)
}
