package comments

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/threadlike/threads"
)

func (GraphQLResolver) CreateComment(ctx context.Context, arg *graphqlbackend.CreateCommentArgs) (graphqlbackend.Comment, error) {
	v := &dbComment{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
		IsPreview:   arg.Input.Preview != nil && *arg.Input.Preview,
	}
	if arg.Input.Rules != nil {
		v.Rules = *arg.Input.Rules
	}

	var err error
	v.NamespaceUserID, v.NamespaceOrgID, err = graphqlbackend.NamespaceDBIDByID(ctx, arg.Input.Namespace)
	if err != nil {
		return nil, err
	}

	comment, err := dbComments{}.Create(ctx, v)
	if err != nil {
		return nil, err
	}
	return &gqlComment{db: comment}, nil
}

func (GraphQLResolver) UpdateComment(ctx context.Context, arg *graphqlbackend.UpdateCommentArgs) (graphqlbackend.Comment, error) {
	l, err := commentByID(ctx, arg.Input.ID)
	if err != nil {
		return nil, err
	}
	comment, err := dbComments{}.Update(ctx, l.db.ID, dbCommentUpdate{
		Name:        arg.Input.Name,
		Description: arg.Input.Description,
		Rules:       arg.Input.Rules,
	})
	if err != nil {
		return nil, err
	}
	return &gqlComment{db: comment}, nil
}

func (GraphQLResolver) PublishPreviewComment(ctx context.Context, arg *graphqlbackend.PublishPreviewCommentArgs) (graphqlbackend.Comment, error) {
	l, err := commentByID(ctx, arg.Comment)
	if err != nil {
		return nil, err
	}

	if !l.IsPreview() {
		return nil, errors.New("comment has already been published (and is not in preview)")
	}

	v := false
	comment, err := dbComments{}.Update(ctx, l.db.ID, dbCommentUpdate{
		IsPreview: &v,
	})
	if err != nil {
		return nil, err
	}
	return &gqlComment{db: comment}, nil
}

func (GraphQLResolver) DeleteComment(ctx context.Context, arg *graphqlbackend.DeleteCommentArgs) (*graphqlbackend.EmptyResponse, error) {
	gqlComment, err := commentByID(ctx, arg.Comment)
	if err != nil {
		return nil, err
	}
	return nil, dbComments{}.DeleteByID(ctx, gqlComment.db.ID)
}

func (GraphQLResolver) AddThreadsToComment(ctx context.Context, arg *graphqlbackend.AddRemoveThreadsToFromCommentArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := addRemoveThreadsToFromComment(ctx, arg.Comment, arg.Threads, nil); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func (GraphQLResolver) RemoveThreadsFromComment(ctx context.Context, arg *graphqlbackend.AddRemoveThreadsToFromCommentArgs) (*graphqlbackend.EmptyResponse, error) {
	if err := addRemoveThreadsToFromComment(ctx, arg.Comment, nil, arg.Threads); err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}

func addRemoveThreadsToFromComment(ctx context.Context, commentID graphql.ID, addThreads []graphql.ID, removeThreads []graphql.ID) error {
	// 🚨 SECURITY: Any viewer can add/remove threads to/from a comment.
	comment, err := commentByID(ctx, commentID)
	if err != nil {
		return err
	}

	if len(addThreads) > 0 {
		addThreadIDs, err := getThreadDBIDs(ctx, addThreads)
		if err != nil {
			return err
		}
		if err := (dbCommentsThreads{}).AddThreadsToComment(ctx, comment.db.ID, addThreadIDs); err != nil {
			return err
		}
	}

	if len(removeThreads) > 0 {
		removeThreadIDs, err := getThreadDBIDs(ctx, removeThreads)
		if err != nil {
			return err
		}
		if err := (dbCommentsThreads{}).RemoveThreadsFromComment(ctx, comment.db.ID, removeThreadIDs); err != nil {
			return err
		}
	}

	return nil
}

var mockGetThreadDBIDs func(threadIDs []graphql.ID) ([]int64, error)

func getThreadDBIDs(ctx context.Context, threadIDs []graphql.ID) ([]int64, error) {
	if mockGetThreadDBIDs != nil {
		return mockGetThreadDBIDs(threadIDs)
	}

	dbIDs := make([]int64, len(threadIDs))
	for i, threadID := range threadIDs {
		// 🚨 SECURITY: Only organization members and site admins may create threads in an
		// organization. The threadByID function performs this check.
		thread, err := threads.GraphQLResolver{}.ThreadByID(ctx, threadID)
		if err != nil {
			return nil, err
		}
		dbIDs[i] = thread.DBID()
	}
	return dbIDs, nil
}
