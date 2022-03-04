package graphql

import (
	"context"
	"fmt"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/internal/squirrel"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type squirrelResolver struct {
	resolver  resolvers.Resolver
	errTracer *observation.ErrCollector
}

func NewSquirrelResolver(resolver resolvers.Resolver, errTracer *observation.ErrCollector) gql.SquirrelResolver {
	return &squirrelResolver{
		resolver:  resolver,
		errTracer: errTracer,
	}
}

func (r *squirrelResolver) Definition(ctx context.Context, args *gql.SquirrelDefinitionArgs) gql.SquirrelLocationResolver {
	result, err := squirrel.DefaultClient.Definition(ctx, args.Location)
	fmt.Println("GQL", result, err)
	if err != nil || result == nil {
		return nil
	}

	return &squirrelLocationResolver{location: *result}
}

type squirrelLocationResolver struct {
	location types.SquirrelLocation
}

func (r *squirrelLocationResolver) Repo() string   { return r.location.Repo }
func (r *squirrelLocationResolver) Commit() string { return r.location.Commit }
func (r *squirrelLocationResolver) Path() string   { return r.location.Path }
func (r *squirrelLocationResolver) Row() int32     { return r.location.Row }
func (r *squirrelLocationResolver) Column() int32  { return r.location.Column }
