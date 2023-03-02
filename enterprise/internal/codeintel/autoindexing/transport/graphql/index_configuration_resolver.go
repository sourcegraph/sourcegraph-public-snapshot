package graphql

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/autoindexing/internal/inference"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/errors"
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

func (r *indexConfigurationResolver) InferredConfiguration(ctx context.Context) (_ resolverstubs.InferredConfigurationResolver, err error) {
	defer r.errTracer.Collect(&err, log.String("indexConfigResolver.field", "inferredConfiguration"))

	var limitErr error
	configuration, _, err := r.autoindexSvc.InferIndexConfiguration(ctx, r.repositoryID, "", "", true)
	if err != nil {
		if errors.As(err, &inference.LimitError{}) {
			limitErr = err
		} else {

			return nil, err
		}
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

	return &inferredConfigurationResolver{
		autoindexSvc:  r.autoindexSvc,
		configuration: indented.String(),
		limitErr:      limitErr,
	}, nil
}

func (r *indexConfigurationResolver) ParsedConfiguration(ctx context.Context) (*[]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	configuration, err := r.Configuration(ctx)
	if err != nil {
		return nil, err
	}
	if configuration == nil {
		return nil, nil
	}

	descriptions, err := newDescriptionResolversFromJSON(r.autoindexSvc, *configuration)
	if err != nil {
		return nil, err
	}

	return &descriptions, nil
}

type inferredConfigurationResolver struct {
	autoindexSvc  AutoIndexingService
	configuration string
	limitErr      error
}

func (r *inferredConfigurationResolver) Configuration() string {
	return r.configuration
}

func (r *inferredConfigurationResolver) ParsedConfiguration(ctx context.Context) (*[]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	descriptions, err := newDescriptionResolversFromJSON(r.autoindexSvc, r.configuration)
	if err != nil {
		return nil, err
	}

	return &descriptions, nil
}

func (r *inferredConfigurationResolver) LimitError() *string {
	if r.limitErr != nil {
		m := r.limitErr.Error()
		return &m
	}

	return nil
}

func newDescriptionResolvers(autoindexSvc AutoIndexingService, indexConfiguration *config.IndexConfiguration) ([]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	var resolvers []resolverstubs.AutoIndexJobDescriptionResolver
	for _, indexJob := range indexConfiguration.IndexJobs {
		var steps []types.DockerStep
		for _, step := range indexJob.Steps {
			steps = append(steps, types.DockerStep{
				Root:     step.Root,
				Image:    step.Image,
				Commands: step.Commands,
			})
		}

		resolvers = append(resolvers, &autoIndexJobDescriptionResolver{
			autoindexSvc: autoindexSvc,
			indexJob:     indexJob,
			steps:        steps,
		})
	}

	return resolvers, nil
}

func newDescriptionResolversFromJSON(autoindexSvc AutoIndexingService, configuration string) ([]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	indexConfiguration, err := config.UnmarshalJSON([]byte(configuration))
	if err != nil {
		return nil, err
	}

	return newDescriptionResolvers(autoindexSvc, &indexConfiguration)
}
