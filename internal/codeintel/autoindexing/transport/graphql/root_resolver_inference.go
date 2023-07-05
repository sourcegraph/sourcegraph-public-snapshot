package graphql

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	sharedresolvers "github.com/sourcegraph/sourcegraph/internal/codeintel/shared/resolvers"
	uploadsshared "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/shared"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
	"github.com/sourcegraph/sourcegraph/lib/pointers"
)

// 🚨 SECURITY: Only site admins may infer auto-index jobs
func (r *rootResolver) InferAutoIndexJobsForRepo(ctx context.Context, args *resolverstubs.InferAutoIndexJobsForRepoArgs) (_ resolverstubs.InferAutoIndexJobsResultResolver, err error) {
	ctx, _, endObservation := r.operations.inferAutoIndexJobsForRepo.WithErrors(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("repository", string(args.Repository)),
		attribute.String("rev", pointers.Deref(args.Rev, "")),
		attribute.String("script", pointers.Deref(args.Script, "")),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := resolverstubs.UnmarshalID[int](args.Repository)
	if err != nil {
		return nil, err
	}

	rev := "HEAD"
	if args.Rev != nil {
		rev = *args.Rev
	}

	localOverrideScript := ""
	if args.Script != nil {
		localOverrideScript = *args.Script
	}

	result, err := r.autoindexSvc.InferIndexConfiguration(ctx, repositoryID, rev, localOverrideScript, false)
	if err != nil {
		return nil, err
	}

	jobResolvers, err := newDescriptionResolvers(r.siteAdminChecker, &config.IndexConfiguration{IndexJobs: result.IndexJobs})
	if err != nil {
		return nil, err
	}

	return &inferAutoIndexJobsResultResolver{
		jobs:            jobResolvers,
		inferenceOutput: result.InferenceOutput,
	}, nil
}

// 🚨 SECURITY: Only site admins may queue auto-index jobs
func (r *rootResolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *resolverstubs.QueueAutoIndexJobsForRepoArgs) (_ []resolverstubs.PreciseIndexResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.queueAutoIndexJobsForRepo.WithErrors(ctx, &err, observation.Args{Attrs: []attribute.KeyValue{
		attribute.String("repository", string(args.Repository)),
		attribute.String("rev", pointers.Deref(args.Rev, "")),
		attribute.String("configuration", pointers.Deref(args.Configuration, "")),
	}})
	endObservation.OnCancel(ctx, 1, observation.Args{})

	if err := r.siteAdminChecker.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		return nil, err
	}
	if !autoIndexingEnabled() {
		return nil, errAutoIndexingNotEnabled
	}

	repositoryID, err := resolverstubs.UnmarshalID[api.RepoID](args.Repository)
	if err != nil {
		return nil, err
	}

	rev := "HEAD"
	if args.Rev != nil {
		rev = *args.Rev
	}

	configuration := ""
	if args.Configuration != nil {
		configuration = *args.Configuration
	}

	indexes, err := r.autoindexSvc.QueueIndexes(ctx, int(repositoryID), rev, configuration, true, true)
	if err != nil {
		return nil, err
	}

	// Create index loader with data we already have
	indexLoader := r.indexLoaderFactory.CreateWithInitialData(indexes)

	// Pre-submit associated upload ids for subsequent loading
	uploadLoader := r.uploadLoaderFactory.Create()
	uploadsgraphql.PresubmitAssociatedUploads(uploadLoader, indexes...)

	// No data to load for git data (yet)
	locationResolver := r.locationResolverFactory.Create()

	resolvers := make([]resolverstubs.PreciseIndexResolver, 0, len(indexes))
	for _, index := range indexes {
		index := index
		resolver, err := r.preciseIndexResolverFactory.Create(ctx, uploadLoader, indexLoader, locationResolver, traceErrs, nil, &index)
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, resolver)
	}

	return resolvers, nil
}

//
//

type inferAutoIndexJobsResultResolver struct {
	jobs            []resolverstubs.AutoIndexJobDescriptionResolver
	inferenceOutput string
}

func (r *inferAutoIndexJobsResultResolver) Jobs() []resolverstubs.AutoIndexJobDescriptionResolver {
	return r.jobs
}

func (r *inferAutoIndexJobsResultResolver) InferenceOutput() string {
	return r.inferenceOutput
}

//
//

type autoIndexJobDescriptionResolver struct {
	siteAdminChecker sharedresolvers.SiteAdminChecker
	indexJob         config.IndexJob
	steps            []uploadsshared.DockerStep
}

func newDescriptionResolvers(siteAdminChecker sharedresolvers.SiteAdminChecker, indexConfiguration *config.IndexConfiguration) ([]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	var resolvers []resolverstubs.AutoIndexJobDescriptionResolver
	for _, indexJob := range indexConfiguration.IndexJobs {
		var steps []uploadsshared.DockerStep
		for _, step := range indexJob.Steps {
			steps = append(steps, uploadsshared.DockerStep{
				Root:     step.Root,
				Image:    step.Image,
				Commands: step.Commands,
			})
		}

		resolvers = append(resolvers, &autoIndexJobDescriptionResolver{
			siteAdminChecker: siteAdminChecker,
			indexJob:         indexJob,
			steps:            steps,
		})
	}

	return resolvers, nil
}

func (r *autoIndexJobDescriptionResolver) Root() string {
	return r.indexJob.Root
}

func (r *autoIndexJobDescriptionResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	return uploadsgraphql.NewCodeIntelIndexerResolver(r.indexJob.Indexer, r.indexJob.Indexer)
}

func (r *autoIndexJobDescriptionResolver) ComparisonKey() string {
	return comparisonKey(r.indexJob.Root, r.Indexer().Name())
}

func (r *autoIndexJobDescriptionResolver) Steps() resolverstubs.IndexStepsResolver {
	return uploadsgraphql.NewIndexStepsResolver(r.siteAdminChecker, uploadsshared.Index{
		DockerSteps:      r.steps,
		LocalSteps:       r.indexJob.LocalSteps,
		Root:             r.indexJob.Root,
		Indexer:          r.indexJob.Indexer,
		IndexerArgs:      r.indexJob.IndexerArgs,
		Outfile:          r.indexJob.Outfile,
		RequestedEnvVars: r.indexJob.RequestedEnvVars,
	})
}

func comparisonKey(root, indexer string) string {
	hash := sha256.New()
	_, _ = hash.Write([]byte(strings.Join([]string{root, indexer}, "\x00")))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}
