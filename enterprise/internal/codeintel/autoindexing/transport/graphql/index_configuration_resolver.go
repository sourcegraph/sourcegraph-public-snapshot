package graphql

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go/log"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

type indexConfigurationResolver struct {
	autoindexSvc AutoIndexingService
	repositoryID int
	errTracer    *observation.ErrCollector
}

func NewIndexConfigurationResolver(autoindexSvc AutoIndexingService, repositoryID int, errTracer *observation.ErrCollector) resolverstubs.IndexConfigurationResolver {
	return &indexConfigurationResolver{
		autoindexSvc: autoindexSvc,
		repositoryID: repositoryID,
		errTracer:    errTracer,
	}
}

func (r *indexConfigurationResolver) Configuration(ctx context.Context) (_ *string, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConfigResolver.field", "configuration"))
	configuration, exists, err := r.autoindexSvc.GetIndexConfigurationByRepositoryID(ctx, r.repositoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return strPtr(string(configuration.Data)), nil
}

func (r *indexConfigurationResolver) InferredConfiguration(ctx context.Context) (_ *string, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConfigResolver.field", "inferredConfiguration"))
	configuration, _, err := r.autoindexSvc.InferIndexConfiguration(ctx, r.repositoryID, "", "", true)
	if err != nil {
		return nil, err
	}
	if configuration == nil {
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
