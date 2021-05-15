package graphql

import (
	"context"
	"encoding/json"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
)

func (r *QueryResolver) DocumentationPage(ctx context.Context, args *gql.LSIFDocumentationPageArgs) (gql.DocumentationPageResolver, error) {
	page, err := r.resolver.DocumentationPage(ctx, args.PathID)
	if err != nil {
		return nil, err
	}
	value, err := json.Marshal(page)
	if err != nil {
		return nil, err
	}
	return &DocumentationPageResolver{tree: gql.JSONValue{Value: string(value)}}, nil
}

type DocumentationPageResolver struct {
	tree gql.JSONValue
}

func (r *DocumentationPageResolver) Tree() gql.JSONValue {
	return r.tree
}
