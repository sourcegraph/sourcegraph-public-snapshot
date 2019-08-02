package threadlike

import (
	"context"
	"errors"
	"path"
	"strconv"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/commentobjectdb"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/internal"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

// GQLThreadlike implements common fields for the GraphQL thread, issue, and changeset types.
type GQLThreadlike struct {
	DB *internal.DBThread
	graphqlbackend.PartialComment
}

func (v *GQLThreadlike) Type() graphqlbackend.ThreadlikeType { return v.DB.Type }

func (v *GQLThreadlike) Repository(ctx context.Context) (*graphqlbackend.RepositoryResolver, error) {
	return graphqlbackend.RepositoryByDBID(ctx, v.DB.RepositoryID)
}

func (v *GQLThreadlike) Number() string { return strconv.FormatInt(v.DB.ID, 10) }

func (v *GQLThreadlike) DBID() int64 { return v.DB.ID }

func (v *GQLThreadlike) Title() string { return v.DB.Title }

func (v *GQLThreadlike) ExternalURL() *string { return v.DB.ExternalURL }

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
	case graphqlbackend.ThreadlikeTypeThread:
		typeComponent = "threads"
	case graphqlbackend.ThreadlikeTypeIssue:
		typeComponent = "issues"
	case graphqlbackend.ThreadlikeTypeChangeset:
		typeComponent = "changesets"
	default:
		return "", errors.New("invalid thread type")
	}

	return path.Join(repository.URL(), "-", typeComponent, v.Number()), nil
}
