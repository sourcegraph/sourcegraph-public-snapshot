package graphqlbackend

import (
	"context"

	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	executor "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
	gql "github.com/sourcegraph/sourcegraph/internal/services/executors/transport/graphql"
)

type ExecutorResolver interface {
	ExecutorResolver() executor.Resolver
}

func (r *schemaResolver) Executors(ctx context.Context, args *struct {
	Query  *string
	Active *bool
	First  *int32
	After  *string
}) (*gql.ExecutorPaginatedResolver, error) {
	// 🚨 SECURITY: Only site-admins may view executor details
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, r.db); err != nil {
		return nil, err
	}

	executors, err := r.ExecutorResolver.ExecutorResolver().Executors(ctx, args.Query, args.Active, args.First, args.After)
	if err != nil {
		return nil, err
	}

	return executors, nil
}

func (r *schemaResolver) AreExecutorsConfigured() bool {
	return conf.Get().ExecutorsAccessToken != ""
}

func executorByID(ctx context.Context, db database.DB, gqlID graphql.ID, r *schemaResolver) (*gql.ExecutorResolver, error) {
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	executor, err := r.ExecutorResolver.ExecutorResolver().Executor(ctx, gqlID)
	if err != nil {
		return nil, err
	}

	return executor, nil
}
