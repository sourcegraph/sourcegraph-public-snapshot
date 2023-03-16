package sharedresolvers

import (
	"context"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/api"
	resolverstubs "github.com/sourcegraph/sourcegraph/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/gqlutil"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadResolver struct {
	uploadsSvc       UploadsService
	policySvc        PolicyService
	gitserverClient  gitserver.Client
	siteAdminChecker SiteAdminChecker
	repoStore        database.RepoStore
	upload           types.Upload
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewUploadResolver(uploadsSvc UploadsService, policySvc PolicyService, gitserverClient gitserver.Client, siteAdminChecker SiteAdminChecker, repoStore database.RepoStore, upload types.Upload, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, traceErrs *observation.ErrCollector) *UploadResolver {
	if upload.AssociatedIndexID != nil {
		// Request the next batch of index fetches to contain the record's associated
		// index id, if one exists it exists. This allows the prefetcher.GetIndexByID
		// invocation in the AssociatedIndex method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkIndex(*upload.AssociatedIndexID)
	}

	return &UploadResolver{
		uploadsSvc:       uploadsSvc,
		policySvc:        policySvc,
		gitserverClient:  gitserverClient,
		siteAdminChecker: siteAdminChecker,
		repoStore:        repoStore,
		upload:           upload,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		traceErrs:        traceErrs,
	}
}

func (r *UploadResolver) ID() graphql.ID        { return marshalLSIFUploadGQLID(int64(r.upload.ID)) }
func (r *UploadResolver) InputCommit() string   { return r.upload.Commit }
func (r *UploadResolver) InputRoot() string     { return r.upload.Root }
func (r *UploadResolver) IsLatestForRepo() bool { return r.upload.VisibleAtTip }
func (r *UploadResolver) UploadedAt() gqlutil.DateTime {
	return gqlutil.DateTime{Time: r.upload.UploadedAt}
}
func (r *UploadResolver) Failure() *string { return r.upload.FailureMessage }
func (r *UploadResolver) StartedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.upload.StartedAt)
}
func (r *UploadResolver) FinishedAt() *gqlutil.DateTime {
	return gqlutil.DateTimeOrNil(r.upload.FinishedAt)
}
func (r *UploadResolver) InputIndexer() string { return r.upload.Indexer }
func (r *UploadResolver) PlaceInQueue() *int32 { return toInt32(r.upload.Rank) }

func (r *UploadResolver) Tags(ctx context.Context) (tagsNames []string, err error) {
	tags, err := r.gitserverClient.ListTags(ctx, api.RepoName(r.upload.RepositoryName), r.upload.Commit)
	if err != nil {
		return nil, err
	}
	for _, tag := range tags {
		tagsNames = append(tagsNames, tag.Name)
	}
	return
}

func (r *UploadResolver) State() string {
	state := strings.ToUpper(r.upload.State)
	if state == "FAILED" {
		state = "ERRORED"
	}

	return state
}

func (r *UploadResolver) AssociatedIndex(ctx context.Context) (_ *indexResolver, err error) {
	// TODO - why are a bunch of them zero?
	if r.upload.AssociatedIndexID == nil || *r.upload.AssociatedIndexID == 0 {
		return nil, nil
	}

	defer r.traceErrs.Collect(&err,
		log.String("uploadResolver.field", "associatedIndex"),
		log.Int("associatedIndex", *r.upload.AssociatedIndexID),
	)

	index, exists, err := r.prefetcher.GetIndexByID(ctx, *r.upload.AssociatedIndexID)
	if err != nil || !exists {
		return nil, err
	}

	return NewIndexResolver(r.uploadsSvc, r.policySvc, r.gitserverClient, r.siteAdminChecker, r.repoStore, index, r.prefetcher, r.locationResolver, r.traceErrs), nil
}

func (r *UploadResolver) ProjectRoot(ctx context.Context) (_ resolverstubs.GitTreeEntryResolver, err error) {
	defer r.traceErrs.Collect(&err, log.String("uploadResolver.field", "projectRoot"))

	resolver, err := r.locationResolver.Path(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root, true)
	if err != nil || resolver == nil {
		// Do not return typed nil interface
		return nil, err
	}

	return resolver, nil
}

const DefaultRetentionPolicyMatchesPageSize = 50

func (r *UploadResolver) RetentionPolicyOverview(ctx context.Context, args *resolverstubs.LSIFUploadRetentionPolicyMatchesArgs) (_ resolverstubs.CodeIntelligenceRetentionPolicyMatchesConnectionResolver, err error) {
	var afterID int64
	if args.After != nil {
		afterID, err = unmarshalConfigurationPolicyGQLID(graphql.ID(*args.After))
		if err != nil {
			return nil, err
		}
	}

	pageSize := DefaultRetentionPolicyMatchesPageSize
	if args.First != nil {
		pageSize = int(*args.First)
	}

	var term string
	if args.Query != nil {
		term = *args.Query
	}

	matches, totalCount, err := r.policySvc.GetRetentionPolicyOverview(ctx, r.upload, args.MatchesOnly, pageSize, afterID, term, time.Now())
	if err != nil {
		return nil, err
	}

	return NewCodeIntelligenceRetentionPolicyMatcherConnectionResolver(r.repoStore, matches, totalCount, r.traceErrs), nil
}

func (r *UploadResolver) Indexer() resolverstubs.CodeIntelIndexerResolver {
	return types.NewCodeIntelIndexerResolver(r.upload.Indexer, "")
}

func (r *UploadResolver) AuditLogs(ctx context.Context) (*[]resolverstubs.LSIFUploadsAuditLogsResolver, error) {
	logs, err := r.uploadsSvc.GetAuditLogsForUpload(ctx, r.upload.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]resolverstubs.LSIFUploadsAuditLogsResolver, 0, len(logs))
	for _, uploadLog := range logs {
		resolvers = append(resolvers, NewLSIFUploadsAuditLogsResolver(uploadLog))
	}

	return &resolvers, nil
}
