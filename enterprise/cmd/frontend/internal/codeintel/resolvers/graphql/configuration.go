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
	errTracer    *observation.ErrCollector
}

func NewIndexConfigurationResolver(resolver resolvers.Resolver, repositoryID int, errTracer *observation.ErrCollector) gql.IndexConfigurationResolver {
	return &IndexConfigurationResolver{
		resolver:     resolver,
		repositoryID: repositoryID,
		errTracer:    errTracer,
	}
}

func (r *IndexConfigurationResolver) Configuration(ctx context.Context) (_ *string, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConfigResolver.field", "configuration"))
	autoIndexingResolver := r.resolver.AutoIndexingResolver()
	configuration, exists, err := autoIndexingResolver.GetIndexConfiguration(ctx, r.repositoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return strPtr(string(configuration)), nil
}

func (r *IndexConfigurationResolver) InferredConfiguration(ctx context.Context) (_ *string, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConfigResolver.field", "inferredConfiguration"))
	autoIndexingResolver := r.resolver.AutoIndexingResolver()
	configuration, exists, err := autoIndexingResolver.InferedIndexConfiguration(ctx, r.repositoryID, "")
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
