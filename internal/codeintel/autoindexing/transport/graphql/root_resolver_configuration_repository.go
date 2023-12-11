package graphql

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/graph-gophers/graphql-go"
	"go.opentelemetry.io/otel/attribute"

	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// 🚨 SECURITY: Only entrypoint is within the repository resolver so the user is already authenticated
func (r *rootResolver) IndexConfiguration(ctx context.Context, repoID graphql.ID) (_ resolverstubs.IndexConfigurationResolver, err error) {
	_, traceErrs, endObservation := r.operations.indexConfiguration.WithErrors(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("repoID", string(repoID)),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := resolverstubs.UnmarshalID[int](repoID)
	if err != nil {
		return nil, err
	}

	return newIndexConfigurationResolver(r.autoindexSvc, r.siteAdminChecker, repositoryID, traceErrs), nil
}

// 🚨 SECURITY: Only site admins may modify code intelligence indexing configuration
func (r *rootResolver) UpdateRepositoryIndexConfiguration(ctx context.Context, args *resolverstubs.UpdateRepositoryIndexConfigurationArgs) (_ *resolverstubs.EmptyResponse, err error) {
	ctx, _, endObservation := r.operations.updateRepositoryIndexConfiguration.With(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("repository", string(args.Repository)),
		attribute.String("configuration", args.Configuration),
	}})
	defer endObservation(1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	// Validate input as JSON
	if _, err := config.UnmarshalJSON([]byte(args.Configuration)); err != nil {
		return nil, err
	}

	repositoryID, err := resolverstubs.UnmarshalID[int](args.Repository)
	if err != nil {
		return nil, err
	}

	if err := r.autoindexSvc.UpdateIndexConfigurationByRepositoryID(ctx, repositoryID, []byte(args.Configuration)); err != nil {
		return nil, err
	}

	return resolverstubs.Empty, nil
}

//
//

type indexConfigurationResolver struct {
	autoindexSvc     AutoIndexingService
	siteAdminChecker sharedresolvers.SiteAdminChecker
	repositoryID     int
	errTracer        *observation.ErrCollector
}

func newIndexConfigurationResolver(autoindexSvc AutoIndexingService, siteAdminChecker sharedresolvers.SiteAdminChecker, repositoryID int, errTracer *observation.ErrCollector) resolverstubs.IndexConfigurationResolver {
	return &indexConfigurationResolver{
		autoindexSvc:     autoindexSvc,
		siteAdminChecker: siteAdminChecker,
		repositoryID:     repositoryID,
		errTracer:        errTracer,
	}
}

func (r *indexConfigurationResolver) Configuration(ctx context.Context) (_ *string, err error) {
	defer r.errTracer.Collect(&err, attribute.String("indexConfigResolver.field", "configuration"))

	configuration, exists, err := r.autoindexSvc.GetIndexConfigurationByRepositoryID(ctx, r.repositoryID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}

	return pointers.NonZeroPtr(string(configuration.Data)), nil
}

func (r *indexConfigurationResolver) InferredConfiguration(ctx context.Context) (_ resolverstubs.InferredConfigurationResolver, err error) {
	defer r.errTracer.Collect(&err, attribute.String("indexConfigResolver.field", "inferredConfiguration"))

	var limitErr error
	result, err := r.autoindexSvc.InferIndexConfiguration(ctx, r.repositoryID, "", "", true)
	if err != nil {
		return nil, err
	}

	marshaled, err := config.MarshalJSON(config.IndexConfiguration{IndexJobs: result.IndexJobs})
	if err != nil {
		return nil, err
	}

	var indented bytes.Buffer
	_ = json.Indent(&indented, marshaled, "", "\t")

	return &inferredConfigurationResolver{
		siteAdminChecker: r.siteAdminChecker,
		configuration:    indented.String(),
		limitErr:         limitErr,
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

	descriptions, err := newDescriptionResolversFromJSON(r.siteAdminChecker, *configuration)
	if err != nil {
		return nil, err
	}

	return &descriptions, nil
}

//
//

type inferredConfigurationResolver struct {
	siteAdminChecker sharedresolvers.SiteAdminChecker
	configuration    string
	limitErr         error
}

func (r *inferredConfigurationResolver) Configuration() string {
	return r.configuration
}

func (r *inferredConfigurationResolver) ParsedConfiguration(ctx context.Context) (*[]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	descriptions, err := newDescriptionResolversFromJSON(r.siteAdminChecker, r.configuration)
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

//
//

func newDescriptionResolversFromJSON(siteAdminChecker sharedresolvers.SiteAdminChecker, configuration string) ([]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	indexConfiguration, err := config.UnmarshalJSON([]byte(configuration))
	if err != nil {
		return nil, err
	}

	return newDescriptionResolvers(siteAdminChecker, &indexConfiguration)
}
