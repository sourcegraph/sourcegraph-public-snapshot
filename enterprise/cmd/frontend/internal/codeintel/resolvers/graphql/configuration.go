package graphql

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type IndexConfigurationResolver struct {
	resolver     resolvers.Resolver
	repositoryID int
	op           *observation.Operation
}

func NewIndexConfigurationResolver(resolver resolvers.Resolver, repositoryID int, op *observation.Operation) gql.IndexConfigurationResolver {
	return &IndexConfigurationResolver{
		resolver:     resolver,
		repositoryID: repositoryID,
		op:           op,
	}
}

func (r *IndexConfigurationResolver) Configuration(ctx context.Context) (_ *string, err error) {
	ctx, endObservatioin := r.op.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("field", "configuration"),
	}})
	defer endObservatioin(1, observation.Args{})

	configuration, exists, err := r.resolver.IndexConfiguration(ctx, r.repositoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return strPtr(string(configuration)), nil
}

func (r *IndexConfigurationResolver) InferredConfiguration(ctx context.Context) (_ *string, err error) {
	ctx, endObservatioin := r.op.With(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("field", "inferredConfiguration"),
	}})
	defer endObservatioin(1, observation.Args{})

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
