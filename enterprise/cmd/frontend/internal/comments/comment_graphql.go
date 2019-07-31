package comments

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

func commentLookupInfoFromGQLID(id graphql.ID) (threadID int64, err error) {
	switch relay.UnmarshalKind(id) {
	case string(threadlike.GQLTypeThread), string(threadlike.GQLTypeIssue), string(threadlike.GQLTypeChangeset):
		_, threadID, err = threadlike.UnmarshalID(id)
	default:
		err = fmt.Errorf("invalid comment type %q", relay.UnmarshalKind(id))
	}
	return
}

func commentByGQLID(ctx context.Context, id graphql.ID) (*dbComment, error) {
	var opt dbCommentsListOptions
	var err error
	opt.ThreadID, err = commentLookupInfoFromGQLID(id)
	if err != nil {
		return nil, err
	}

	comments, err := dbComments{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return nil, errCommentNotFound
	}
	if len(comments) >= 2 {
		return nil, fmt.Errorf("got %d comments for GraphQL ID %q, expected 0 or 1", len(comments), id)
	}
	return comments[0], nil
}

func newGQLToComment(ctx context.Context, dbComment *dbComment) (*graphqlbackend.ToComment, error) {
	switch {
	case dbComment.ThreadID != 0:
		v, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, dbComment.ThreadID)
		if err != nil {
			return nil, err
		}
		switch {
		case v.Thread != nil:
			return &graphqlbackend.ToComment{Thread: v.Thread}, nil
			// TODO!(sqs): add Issue switch-case branch
		case v.Changeset != nil:
			return &graphqlbackend.ToComment{Changeset: v.Changeset}, nil
		}
	}
	return nil, errors.New("invalid comment")
}

// GQLIComment implements the GraphQL interface Comment.
type GQLIComment struct{ db *dbComment }

func (v *GQLIComment) Author(ctx context.Context) (*graphqlbackend.Actor, error) {
	user, err := graphqlbackend.UserByIDInt32(ctx, v.db.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.Actor{User: user}, nil
}

func (v *GQLIComment) Body() string { return v.db.Body }

func (v *GQLIComment) CreatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{v.db.CreatedAt}
}
func (v *GQLIComment) UpdatedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{v.db.UpdatedAt}
}
