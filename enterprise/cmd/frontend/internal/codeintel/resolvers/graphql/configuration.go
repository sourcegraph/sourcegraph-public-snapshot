package graphql

import (
	"bytes"
	"context"
	"encoding/json"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type IndexConfigurationResolver struct {
	resolver     resolvers.Resolver
	repositoryID int
}

func NewIndexConfigurationResolver(resolver resolvers.Resolver, repositoryID int) gql.IndexConfigurationResolver {
	return &IndexConfigurationResolver{
		resolver:     resolver,
		repositoryID: repositoryID,
	}
}

func (r *IndexConfigurationResolver) Configuration(ctx context.Context) (*string, error) {
	configuration, exists, err := r.resolver.IndexConfiguration(ctx, r.repositoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return strPtr(string(configuration)), nil
}

func (r *IndexConfigurationResolver) InferredConfiguration(ctx context.Context) (*string, error) {
	configuration, exists, err := r.resolver.InferredIndexConfiguration(ctx, r.repositoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	marshaled, err := config.MarshalJSON(*configuration)
	if err != nil {
		return nil, err
	}

	var indented bytes.Buffer
	_ = json.Indent(&indented, marshaled, "", "\t")

	return strPtr(indented.String()), nil
}
