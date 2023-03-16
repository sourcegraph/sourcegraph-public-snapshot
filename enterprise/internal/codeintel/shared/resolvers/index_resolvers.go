package sharedresolvers

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gitserver/gitdomain"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type indexResolver struct {
	uploadsSvc       UploadsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker SiteAdminChecker
	repoStore        database.RepoStore
	index            types.Index
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewIndexResolver(uploadsSvc UploadsService, policySvc PolicyService, gitserverClient gitserver.Client, siteAdminChecker SiteAdminChecker, repoStore database.RepoStore, index types.Index, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, errTrace *observation.ErrCollector) *indexResolver {
	if index.AssociatedUploadID != nil {
		// Request the next batch of upload fetches to contain the record's associated
		// upload id, if one exists it exists. This allows the prefetcher.GetUploadByID
		// invocation in the AssociatedUpload method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkUpload(*index.AssociatedUploadID)
	}

	return &indexResolver{
		uploadsSvc:       uploadsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		index:            index,
		repoStore:        repoStore,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		traceErrs:        errTrace,
	}
}

func (r *indexResolver) ID() graphql.ID             { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *indexResolver) InputCommit() string        { return r.index.Commit }
func (r *indexResolver) InputRoot() string          { return r.index.Root }
func (r *indexResolver) InputIndexer() string       { return r.index.Indexer }
func (r *indexResolver) QueuedAt() gqlutil.DateTime { return gqlutil.DateTime{Time: r.index.QueuedAt} }
func (r *indexResolver) Failure() *string           { return r.index.FailureMessage }
func (r *indexResolver) StartedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.index.StartedAt)
}

func (r *indexResolver) FinishedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.index.FinishedAt)
}

func (r *indexResolver) Steps() resolverstubs.IndexStepsResolver {
	return NewIndexStepsResolver(r.siteAdminChecker, r.index)
}
func (r *indexResolver) PlaceInQueue() *int32 { return toInt32(r.index.Rank) }

func (r *indexResolver) Tags(ctx context.Context) (tagsNames []string, err error) {
	tags, err := r.gitserverClient.ListTags(ctx, api.RepoName(r.index.RepositoryName), r.index.Commit)
	if err != nil {
		if gitdomain.IsRepoNotExist(err) {
			return tagsNames, nil
		}
		return nil, errors.New("unable to return list of tags in the repository.")
	}
	for _, tag := range tags {
		tagsNames = append(tagsNames, tag.Name)
	}
	return
}

func (r *indexResolver) State() string {
	state := strings.ToUpper(r.index.State)
	if state == "FAILED" {
		state = "ERRORED"
	}

	return state
}

func (r *indexResolver) AssociatedUpload(ctx context.Context) (_ *UploadResolver, err error) {
	if r.index.AssociatedUploadID == nil {
		return nil, nil
	}

	defer r.traceErrs.Collect(&err,
		log.String("indexResolver.field", "associatedUpload"),
		log.Int("associatedUpload", *r.index.AssociatedUploadID),
	)

	upload, exists, err := r.prefetcher.GetUploadByID(ctx, *r.index.AssociatedUploadID)
	if err != nil || !exists {
		return nil, err
	}

	return NewUploadResolver(r.uploadsSvc, r.policySvc, r.gitserverClient, r.siteAdminChecker, r.repoStore, upload, r.prefetcher, r.locationResolver, r.traceErrs), nil
}

func (r *indexResolver) ShouldReindex(ctx context.Context) bool {
	return r.index.ShouldReindex
}

func (r *indexResolver) ProjectRoot(ctx context.Context) (_ resolverstubs.GitTreeEntryResolver, err error) {
	defer r.traceErrs.Collect(&err, log.String("indexResolver.field", "projectRoot"))

	resolver, err := r.locationResolver.Path(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, r.index.Root, true)
	if err != nil || resolver == nil {
		// Do not return typed nil interface
		return nil, err
	}

	return resolver, nil

}

func (r *indexResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	return types.NewCodeIntelIndexerResolver(r.index.Indexer, r.index.Indexer)
}
