package resolvers

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/notebooks"
)

func marshalNotebookStarCursor(cursor int64) string {
	return string(relay.MarshalID("NotebookStarCursor", cursor))
}

func unmarshalNotebookStarCursor(cursor *string) (int64, error) {
	if cursor == nil {
		return 0, nil
	}
	var after int64
	err := relay.UnmarshalSpec(graphql.ID(*cursor), &after)
	if err != nil {
		return -1, err
	}
	return after, nil
}

type notebookStarConnectionResolver struct {
	afterCursor int64
	stars       []graphqlbackend.NotebookStarResolver
	totalCount  int32
	hasNextPage bool
}

func (n *notebookStarConnectionResolver) Nodes() []graphqlbackend.NotebookStarResolver {
	return n.stars
}

func (n *notebookStarConnectionResolver) TotalCount() int32 {
	return n.totalCount
}

func (n *notebookStarConnectionResolver) PageInfo() *gqlutil.PageInfo {
	if len(n.stars) == 0 || !n.hasNextPage {
		return gqlutil.HasNextPage(false)
	}
	// The after value (offset) for the next page is computed from the current after value + the number of retrieved notebook stars
	return gqlutil.NextPageCursor(marshalNotebookStarCursor(n.afterCursor + int64(len(n.stars))))
}

type notebookStarResolver struct {
	star *notebooks.NotebookStar
	db   database.DB
}

func (r *notebookStarResolver) User(ctx context.Context) (*graphqlbackend.UserResolver, error) {
	return graphqlbackend.UserByIDInt32(ctx, r.db, r.star.UserID)
}

func (r *notebookStarResolver) CreatedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.star.CreatedAt}
}

func (r *notebookResolver) notebookStarsToResolvers(notebookStars []*notebooks.NotebookStar) []graphqlbackend.NotebookStarResolver {
	notebookStarsResolvers := make([]graphqlbackend.NotebookStarResolver, len(notebookStars))
	for idx, star := range notebookStars {
		notebookStarsResolvers[idx] = &notebookStarResolver{star, r.db}
	}
	return notebookStarsResolvers
}

func (r *notebookResolver) Stars(ctx context.Context, args graphqlbackend.ListNotebookStarsArgs) (graphqlbackend.NotebookStarConnectionResolver, error) {
	// Request one extra to determine if there are more pages
	newArgs := args
	newArgs.First += 1

	afterCursor, err := unmarshalNotebookStarCursor(args.After)
	if err != nil {
		return nil, err
	}

	pageOpts := notebooks.ListNotebookStarsPageOptions{First: newArgs.First, After: afterCursor}
	store := notebooks.Notebooks(r.db)
	stars, err := store.ListNotebookStars(ctx, pageOpts, r.notebook.ID)
	if err != nil {
		return nil, err
	}

	count, err := store.CountNotebookStars(ctx, r.notebook.ID)
	if err != nil {
		return nil, err
	}

	hasNextPage := false
	if len(stars) == int(args.First)+1 {
		hasNextPage = true
		stars = stars[:len(stars)-1]
	}

	return &notebookStarConnectionResolver{
		afterCursor: afterCursor,
		stars:       r.notebookStarsToResolvers(stars),
		totalCount:  int32(count),
		hasNextPage: hasNextPage,
	}, nil
}

func (r *Resolver) CreateNotebookStar(ctx context.Context, args graphqlbackend.CreateNotebookStarInputArgs) (graphqlbackend.NotebookStarResolver, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	notebookID, err := unmarshalNotebookID(args.NotebookID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	// Ensure user has access to the notebook.
	notebook, err := store.GetNotebook(ctx, notebookID)
	if err != nil {
		return nil, err
	}

	createdStar, err := store.CreateNotebookStar(ctx, notebook.ID, user.ID)
	if err != nil {
		return nil, err
	}
	return &notebookStarResolver{createdStar, r.db}, nil
}

func (r *Resolver) DeleteNotebookStar(ctx context.Context, args graphqlbackend.DeleteNotebookStarInputArgs) (*graphqlbackend.EmptyResponse, error) {
	user, err := r.db.Users().GetByCurrentAuthUser(ctx)
	if err != nil {
		return nil, err
	}

	notebookID, err := unmarshalNotebookID(args.NotebookID)
	if err != nil {
		return nil, err
	}

	store := notebooks.Notebooks(r.db)
	// Ensure user has access to the notebook.
	notebook, err := store.GetNotebook(ctx, notebookID)
	if err != nil {
		return nil, err
	}

	err = store.DeleteNotebookStar(ctx, notebook.ID, user.ID)
	if err != nil {
		return nil, err
	}
	return &graphqlbackend.EmptyResponse{}, nil
}
