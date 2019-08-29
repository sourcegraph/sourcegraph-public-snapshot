package comments

import (
	"context"
	"fmt"
	"html"
	"sync"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/microcosm-cc/bluemonday"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/internal"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/comments/types"
	"github.com/sourcegraph/sourcegraph/pkg/markdown"
)

// 🚨 SECURITY: TODO!(sqs): there needs to be security checks everywhere here! there are none

func commentObjectFromGQLID(id graphql.ID) (types.CommentObject, error) {
	switch relay.UnmarshalKind(id) {
	case gqlTypeCommentReply:
		panic("CommentReply is its own comment object") // TODO!(sqs): clean up
	case "Thread":
		threadID, err := graphqlbackend.UnmarshalThreadID(id)
		return types.CommentObject{ThreadID: threadID}, err
	case "Campaign":
		// TODO!(sqs): reduce duplication of logic and constants?
		var dbID int64
		err := relay.UnmarshalSpec(id, &dbID)
		return types.CommentObject{CampaignID: dbID}, err
	default:
		return types.CommentObject{}, fmt.Errorf("invalid comment type %q", relay.UnmarshalKind(id))
	}
}

var mockCommentByGQLID func(graphql.ID) (*internal.DBComment, error)

func commentByGQLID(ctx context.Context, id graphql.ID) (*internal.DBComment, error) {
	if mockCommentByGQLID != nil {
		return mockCommentByGQLID(id)
	}

	// Look up a CommentReply directly because its ID directly refers to its comment.
	if relay.UnmarshalKind(id) == gqlTypeCommentReply {
		dbID, err := unmarshalCommentReplyID(id)
		if err != nil {
			return nil, err
		}
		return internal.DBComments{}.GetByID(ctx, dbID)
	}

	opt := internal.DBCommentsListOptions{ObjectPrimaryComment: true}
	var err error
	opt.Object, err = commentObjectFromGQLID(id)
	if err != nil {
		return nil, err
	}
	comments, err := internal.DBComments{}.List(ctx, opt)
	if err != nil {
		return nil, err
	}
	if len(comments) == 0 {
		return nil, internal.ErrCommentNotFound
	}
	if len(comments) >= 2 {
		return nil, fmt.Errorf("got %d comments for GraphQL ID %q, expected 0 or 1", len(comments), id)
	}
	return comments[0], nil
}

var mockNewGQLToComment func(*internal.DBComment) (graphqlbackend.Comment, error)

func newGQLToComment(ctx context.Context, dbComment *internal.DBComment) (graphqlbackend.Comment, error) {
	if mockNewGQLToComment != nil {
		return mockNewGQLToComment(dbComment)
	}

	switch {
	case dbComment.Object.ParentCommentID != 0:
		return &graphqlbackend.ToComment{
			CommentReply: &gqlCommentReply{
				gqlComment: &gqlComment{dbComment: dbComment},
			},
		}, nil
	case dbComment.Object.ThreadID != 0:
		v, err := graphqlbackend.ThreadByID(ctx, graphqlbackend.MarshalThreadID(dbComment.Object.ThreadID))
		if err != nil {
			return nil, err
		}
		return &graphqlbackend.ToComment{Thread: v}, nil
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
	dbComment *internal.DBComment
	err       error
}

func (v *gqlComment) getComment(ctx context.Context) (*internal.DBComment, error) {
	v.once.Do(func() {
		if v.dbComment != nil {
			// The dbComment was already present when the struct was instantiated; no need to query
			// the DB.
			return
		}
		v.dbComment, v.err = commentByGQLID(ctx, v.id)
	})
	return v.dbComment, v.err
}

func (v *gqlComment) Author(ctx context.Context) (*graphqlbackend.Actor, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return nil, err
	}
	return c.Author.GQL(ctx)
}

func (v *gqlComment) Body(ctx context.Context) (string, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return "", err
	}
	return c.Body, nil
}

func ToBodyText(body string) string {
	// TODO!(sqs): this doesnt remove markdown formatting like `*`, just HTML tags
	return html.UnescapeString(bluemonday.StrictPolicy().Sanitize(body))
}

func (v *gqlComment) BodyText(ctx context.Context) (string, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return "", err
	}
	return ToBodyText(c.Body), nil
}

func ToBodyHTML(body string) string {
	return markdown.Render(body, nil)
}

func (v *gqlComment) BodyHTML(ctx context.Context) (string, error) {
	c, err := v.getComment(ctx)
	if err != nil {
		return "", err
	}
	return ToBodyHTML(c.Body), nil
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

var DBGetByID = (internal.DBComments{}).GetByID
