package graphql

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"github.com/opentracing/opentracing-go/log"

	sharedresolvers "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/resolvers"
	codeinteltypes "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	uploadsgraphql "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/uploads/transport/graphql"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/codeintel/autoindex/config"
)

// 🚨 SECURITY: Only site admins may infer auto-index jobs
func (r *rootResolver) InferAutoIndexJobsForRepo(ctx context.Context, args *resolverstubs.InferAutoIndexJobsForRepoArgs) (_ []resolverstubs.AutoIndexJobDescriptionResolver, err error) {
	ctx, _, endObservation := r.operations.inferAutoIndexJobsForRepo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repository", string(args.Repository)),
		log.String("rev", resolverstubs.Deref(args.Rev, "")),
		log.String("script", resolverstubs.Deref(args.Script, "")),
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

	// TODO - expose hints
	config, _, err := r.autoindexSvc.InferIndexConfiguration(ctx, repositoryID, rev, localOverrideScript, false)
	if err != nil {
		return nil, err
	}

	if config == nil {
		return nil, nil
	}

	return newDescriptionResolvers(r.siteAdminChecker, config)
}

// 🚨 SECURITY: Only site admins may queue auto-index jobs
func (r *rootResolver) QueueAutoIndexJobsForRepo(ctx context.Context, args *resolverstubs.QueueAutoIndexJobsForRepoArgs) (_ []resolverstubs.PreciseIndexResolver, err error) {
	ctx, traceErrs, endObservation := r.operations.queueAutoIndexJobsForRepo.WithErrors(ctx, &err, observation.Args{LogFields: []log.Field{
		log.String("repository", string(args.Repository)),
		log.String("rev", resolverstubs.Deref(args.Rev, "")),
		log.String("configuration", resolverstubs.Deref(args.Configuration, "")),
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

	prefetcher := r.prefetcherFactory.Create()

	for _, index := range indexes {
		prefetcher.MarkIndex(index.ID)
	}

	resolvers := make([]resolverstubs.PreciseIndexResolver, 0, len(indexes))
	for _, index := range indexes {
		index := index
		resolver, err := uploadsgraphql.NewPreciseIndexResolver(ctx, r.uploadSvc, r.policySvc, r.gitserverClient, prefetcher, r.siteAdminChecker, r.repoStore, r.locationResolverFactory.Create(), traceErrs, nil, &index)
		if err != nil {
			return nil, err
		}

		resolvers = append(resolvers, resolver)
	}

	return resolvers, nil
}

//
//

type autoIndexJobDescriptionResolver struct {
	siteAdminChecker sharedresolvers.SiteAdminChecker
	indexJob         config.IndexJob
	steps            []codeinteltypes.DockerStep
}

func newDescriptionResolvers(siteAdminChecker sharedresolvers.SiteAdminChecker, indexConfiguration *config.IndexConfiguration) ([]resolverstubs.AutoIndexJobDescriptionResolver, error) {
	var resolvers []resolverstubs.AutoIndexJobDescriptionResolver
	for _, indexJob := range indexConfiguration.IndexJobs {
		var steps []codeinteltypes.DockerStep
		for _, step := range indexJob.Steps {
			steps = append(steps, codeinteltypes.DockerStep{
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
	return codeinteltypes.NewCodeIntelIndexerResolver(r.indexJob.Indexer, r.indexJob.Indexer)
}

func (r *autoIndexJobDescriptionResolver) ComparisonKey() string {
	return comparisonKey(r.indexJob.Root, r.Indexer().Name())
}

func (r *autoIndexJobDescriptionResolver) Steps() resolverstubs.IndexStepsResolver {
	return uploadsgraphql.NewIndexStepsResolver(r.siteAdminChecker, codeinteltypes.Index{
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
