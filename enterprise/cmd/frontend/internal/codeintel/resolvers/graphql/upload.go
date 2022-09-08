package graphql

import (
	"context"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type UploadResolver struct {
	db               database.DB
	gitserver        GitserverClient
	resolver         resolvers.Resolver
	upload           dbstore.Upload
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewUploadResolver(db database.DB, gitserver GitserverClient, resolver resolvers.Resolver, upload dbstore.Upload, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, traceErrs *observation.ErrCollector) gql.LSIFUploadResolver {
	if upload.AssociatedIndexID != nil {
		// Request the next batch of index fetches to contain the record's associated
		// index id, if one exists it exists. This allows the prefetcher.GetIndexByID
		// invocation in the AssociatedIndex method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkIndex(*upload.AssociatedIndexID)
	}

	return &UploadResolver{
		db:               db,
		gitserver:        gitserver,
		resolver:         resolver,
		upload:           upload,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		traceErrs:        traceErrs,
	}
}

func (r *UploadResolver) ID() graphql.ID            { return marshalLSIFUploadGQLID(int64(r.upload.ID)) }
func (r *UploadResolver) InputCommit() string       { return r.upload.Commit }
func (r *UploadResolver) InputRoot() string         { return r.upload.Root }
func (r *UploadResolver) IsLatestForRepo() bool     { return r.upload.VisibleAtTip }
func (r *UploadResolver) UploadedAt() gql.DateTime  { return gql.DateTime{Time: r.upload.UploadedAt} }
func (r *UploadResolver) Failure() *string          { return r.upload.FailureMessage }
func (r *UploadResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.upload.StartedAt) }
func (r *UploadResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.upload.FinishedAt) }
func (r *UploadResolver) InputIndexer() string      { return r.upload.Indexer }
func (r *UploadResolver) PlaceInQueue() *int32      { return toInt32(r.upload.Rank) }

func (r *UploadResolver) Tags(ctx context.Context) (tagsNames []string, err error) {
	tags, err := r.gitserver.ListTags(ctx, api.RepoName(r.upload.RepositoryName), r.upload.Commit)
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

func (r *UploadResolver) AssociatedIndex(ctx context.Context) (_ gql.LSIFIndexResolver, err error) {
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

	return NewIndexResolver(r.db, r.gitserver, r.resolver, index, r.prefetcher, r.locationResolver, r.traceErrs), nil
}

func (r *UploadResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return r.locationResolver.Path(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root)
}

func (r *UploadResolver) RetentionPolicyOverview(ctx context.Context, args *gql.LSIFUploadRetentionPolicyMatchesArgs) (_ gql.CodeIntelligenceRetentionPolicyMatchesConnectionResolver, err error) {
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

	policyResolver, err := r.resolver.PoliciesResolver().PolicyResolverFactory(ctx)
	if err != nil {
		return nil, err
	}

	upload := sharedPoliciesUploadsToStoreUpload(r.upload)
	m, totalCount, err := policyResolver.GetRetentionPolicyOverview(ctx, upload, args.MatchesOnly, pageSize, afterID, term, time.Now())
	if err != nil {
		return nil, err
	}
	matches := sharedRetentionPolicyToStoreRetentionPolicy(m)

	return NewCodeIntelligenceRetentionPolicyMatcherConnectionResolver(r.db, r.resolver, matches, totalCount, r.traceErrs), nil
}

func (r *UploadResolver) Indexer() gql.CodeIntelIndexerResolver {
	for _, indexer := range allIndexers {
		if indexer.Name() == r.upload.Indexer {
			return indexer
		}
	}

	return &codeIntelIndexerResolver{name: r.upload.Indexer}
}

func (r *UploadResolver) DocumentPaths(ctx context.Context, args *gql.LSIFUploadDocumentPathsQueryArgs) (gql.LSIFUploadDocumentPathsConnectionResolver, error) {
	pattern := "%%"
	if args.Pattern != "" {
		pattern = args.Pattern
	}

	documents, totalCount, err := r.resolver.UploadsResolver().GetUploadDocumentsForPath(ctx, r.upload.ID, pattern)
	if err != nil {
		return nil, err
	}

	return &uploadDocumentPathsConnectionResolver{
		totalCount: totalCount,
		documents:  documents,
	}, nil
}

func (r *UploadResolver) AuditLogs(ctx context.Context) (*[]gql.LSIFUploadsAuditLogsResolver, error) {
	logs, err := r.resolver.UploadsResolver().GetAuditLogsForUpload(ctx, r.upload.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]gql.LSIFUploadsAuditLogsResolver, 0, len(logs))
	for _, log := range logs {
		resolvers = append(resolvers, &lsifUploadsAuditLogResolver{log})
	}

	return &resolvers, nil
}
