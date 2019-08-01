package comments

import (
	"context"
	"fmt"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

func commentObjectFromGQLID(id graphql.ID) (dbCommentObject, error) {
	switch relay.UnmarshalKind(id) {
	case string(threadlike.GQLTypeThread), string(threadlike.GQLTypeIssue), string(threadlike.GQLTypeChangeset):
		_, threadID, err := threadlike.UnmarshalID(id)
		return dbCommentObject{ThreadID: threadID}, err
	case "Campaign":
		// TODO!(sqs): reduce duplication of logic and constants?
		var dbID int64
		err := relay.UnmarshalSpec(id, &dbID)
		return dbCommentObject{CampaignID: dbID}, err
	default:
		return dbCommentObject{}, fmt.Errorf("invalid comment type %q", relay.UnmarshalKind(id))
	}
}

func commentByGQLID(ctx context.Context, id graphql.ID) (*dbComment, error) {
	if mocks.commentByGQLID != nil {
		return mocks.commentByGQLID(id)
	}

	var opt dbCommentsListOptions
	var err error
	opt.Object, err = commentObjectFromGQLID(id)
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

func newGQLToComment(ctx context.Context, dbComment *dbComment) (graphqlbackend.Comment, error) {
	if mocks.newGQLToComment != nil {
		return mocks.newGQLToComment(dbComment)
	}

	switch {
	case dbComment.Object.ThreadID != 0:
		v, err := threadlike.ThreadOrIssueOrChangesetByDBID(ctx, dbComment.Object.ThreadID)
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
	case dbComment.Object.CampaignID != 0:
		v, err := graphqlbackend.CampaignByDBID(ctx, dbComment.Object.CampaignID)
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.ToComment{Campaign: v}, nil
	}
	return nil, errors.New("invalid comment")
}

func (GraphQLResolver) LazyCommentByID(object graphql.ID) graphqlbackend.PartialComment {
	return &gqlComment{id: object}
}

// gqlComment implements the GraphQL interface Comment.
type gqlComment struct {
	id graphql.ID

	once      sync.Once
	dbComment *dbComment
	err       error
}

// TODO!(sqs) remove
// func (gqlComment) ID() graphql.ID {
// 	panic("The (gqlComment).ID method should not be called. If gqlComment is embedded in another struct type, the struct type must define its own ID method that shadows (gqlComment).ID. This is because gqlComment is lazy and the ID is not guaranteed to be valid when gqlComment being instantiated, so the embedding type may report incorrect IDs to GraphQL API consumers.")
// }

func (v *gqlComment) getComment(ctx context.Context) (*dbComment, error) {
	v.once.Do(func() {
		v.dbComment, v.err = commentByGQLID(ctx, v.id)
	})
	return v.dbComment, v.err
}

func (v *gqlComment) Author(ctx context.Context) (*graphqlbackend.Actor, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return nil, err
	}

	user, err := graphqlbackend.UserByIDInt32(ctx, c.AuthorUserID)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.Actor{User: user}, nil
}

func (v *gqlComment) Body(ctx context.Context) (string, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return "", err
	}
	return c.Body, nil
}

func (v *gqlComment) BodyHTML(ctx context.Context) (string, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return "", err
	}
	return markdown.Render(c.Body, nil), nil
}

func (v *gqlComment) CreatedAt(ctx context.Context) (graphqlbackend.DateTime, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return graphqlbackend.DateTime{}, err
	}
	return graphqlbackend.DateTime{c.CreatedAt}, nil
}

func (v *gqlComment) UpdatedAt(ctx context.Context) (graphqlbackend.DateTime, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return graphqlbackend.DateTime{}, err
	}
	return graphqlbackend.DateTime{c.UpdatedAt}, nil
}
