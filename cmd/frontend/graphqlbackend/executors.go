package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/services/executors"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/services/executors/graphql"
	"github.com/sourcegraph/sourcegraph/internal/database"
)

const DefaultExecutorsLimit = 50

func (r *schemaResolver) Executors(ctx context.Context, args *struct {
	Query  *string
	Active *bool
	First  *int32
	After  *string
}) (*gql.ExecutorPaginatedConnection, error) {
	// 🚨 SECURITY: Only site-admins may view executor details
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	query, active, offset, limit, err := validateArgs(ctx, args)
	if err != nil {
		return nil, err
	}

	executorService := executors.New(r.db)
	executors, totalCount, err := executorService.List(ctx, query, active, offset, limit)
	if err != nil {
		return nil, err
	}

	return executorService.ToPaginatedConnection(ctx, executors, totalCount, offset), nil
}

func executorByID(ctx context.Context, db database.DB, gqlID graphql.ID) (*gql.ExecutorResolver, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	executorService := executors.New(db)
	executor, ok, err := executorService.GetByID(ctx, gqlID)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, nil
	}

	return executorService.ToResolver(ctx, executor), nil
}

type graphqlArgs *struct {
	Query  *string
	Active *bool
	First  *int32
	After  *string
}

func validateArgs(ctx context.Context, args graphqlArgs) (query string, active bool, offset int, limit int, err error) {
	if args.Query != nil {
		query = *args.Query
	}

	if args.Active != nil {
		active = *args.Active
	}

	offset, err = graphqlutil.DecodeIntCursor(args.After)
	if err != nil {
		return
	}

	limit = DefaultExecutorsLimit
	if args.First != nil {
		limit = int(*args.First)
	}

	return
}
