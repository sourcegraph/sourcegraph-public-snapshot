package sharedresolvers

import (
	"context"
	"strings"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/shared/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type LSIFUploadResolver interface {
	ID() graphql.ID
	InputCommit() string
	Tags(ctx context.Context) ([]string, error)
	InputRoot() string
	IsLatestForRepo() bool
	UploadedAt() DateTime
	State() string
	Failure() *string
	StartedAt() *DateTime
	FinishedAt() *DateTime
	InputIndexer() string
	Indexer() types.CodeIntelIndexerResolver
	PlaceInQueue() *int32
	AssociatedIndex(ctx context.Context) (LSIFIndexResolver, error)
	ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error)
	RetentionPolicyOverview(ctx context.Context, args *LSIFUploadRetentionPolicyMatchesArgs) (CodeIntelligenceRetentionPolicyMatchesConnectionResolver, error)
	DocumentPaths(ctx context.Context, args *LSIFUploadDocumentPathsQueryArgs) (LSIFUploadDocumentPathsConnectionResolver, error)
	AuditLogs(ctx context.Context) (*[]LSIFUploadsAuditLogsResolver, error)
}

type UploadResolver struct {
	uploadsSvc       UploadsService
	autoindexingSvc  AutoIndexingService
	policySvc        PolicyService
	upload           types.Upload
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewUploadResolver(uploadsSvc UploadsService, autoindexingSvc AutoIndexingService, policySvc PolicyService, upload types.Upload, prefetcher *Prefetcher, traceErrs *observation.ErrCollector) LSIFUploadResolver {
	if upload.AssociatedIndexID != nil {
		// Request the next batch of index fetches to contain the record's associated
		// index id, if one exists it exists. This allows the prefetcher.GetIndexByID
		// invocation in the AssociatedIndex method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkIndex(*upload.AssociatedIndexID)
	}

	return &UploadResolver{
		uploadsSvc:       uploadsSvc,
		autoindexingSvc:  autoindexingSvc,
		policySvc:        policySvc,
		upload:           upload,
		prefetcher:       prefetcher,
		locationResolver: NewCachedLocationResolver(autoindexingSvc.GetUnsafeDB()),
		traceErrs:        traceErrs,
	}
}

func (r *UploadResolver) ID() graphql.ID        { return marshalLSIFUploadGQLID(int64(r.upload.ID)) }
func (r *UploadResolver) InputCommit() string   { return r.upload.Commit }
func (r *UploadResolver) InputRoot() string     { return r.upload.Root }
func (r *UploadResolver) IsLatestForRepo() bool { return r.upload.VisibleAtTip }
func (r *UploadResolver) UploadedAt() DateTime  { return DateTime{Time: r.upload.UploadedAt} }
func (r *UploadResolver) Failure() *string      { return r.upload.FailureMessage }
func (r *UploadResolver) StartedAt() *DateTime  { return DateTimeOrNil(r.upload.StartedAt) }
func (r *UploadResolver) FinishedAt() *DateTime { return DateTimeOrNil(r.upload.FinishedAt) }
func (r *UploadResolver) InputIndexer() string  { return r.upload.Indexer }
func (r *UploadResolver) PlaceInQueue() *int32  { return toInt32(r.upload.Rank) }

func (r *UploadResolver) Tags(ctx context.Context) (tagsNames []string, err error) {
	tags, err := r.uploadsSvc.GetListTags(ctx, api.RepoName(r.upload.RepositoryName), r.upload.Commit)
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

func (r *UploadResolver) AssociatedIndex(ctx context.Context) (_ LSIFIndexResolver, err error) {
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

	return NewIndexResolver(r.autoindexingSvc, r.uploadsSvc, r.policySvc, index, r.prefetcher, r.traceErrs), nil
}

func (r *UploadResolver) ProjectRoot(ctx context.Context) (*GitTreeEntryResolver, error) {
	return r.locationResolver.Path(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root)
}

const DefaultRetentionPolicyMatchesPageSize = 50

type LSIFUploadRetentionPolicyMatchesArgs struct {
	MatchesOnly bool
	First       *int32
	After       *string
	Query       *string
}

func (r *UploadResolver) RetentionPolicyOverview(ctx context.Context, args *LSIFUploadRetentionPolicyMatchesArgs) (_ CodeIntelligenceRetentionPolicyMatchesConnectionResolver, err error) {
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

	return NewCodeIntelligenceRetentionPolicyMatcherConnectionResolver(r.autoindexingSvc, matches, totalCount, r.traceErrs), nil
}

func (r *UploadResolver) Indexer() types.CodeIntelIndexerResolver {
	for _, indexer := range types.AllIndexers {
		if indexer.Name == r.upload.Indexer {
			return types.NewCodeIntelIndexerResolverFrom(indexer)
		}
	}

	return types.NewCodeIntelIndexerResolver(r.upload.Indexer)
}

type LSIFUploadDocumentPathsQueryArgs struct {
	Pattern string
}

func (r *UploadResolver) DocumentPaths(ctx context.Context, args *LSIFUploadDocumentPathsQueryArgs) (LSIFUploadDocumentPathsConnectionResolver, error) {
	pattern := "%%"
	if args.Pattern != "" {
		pattern = args.Pattern
	}

	documents, totalCount, err := r.uploadsSvc.GetUploadDocumentsForPath(ctx, r.upload.ID, pattern)
	if err != nil {
		return nil, err
	}

	return NewUploadDocumentPathsConnectionResolver(totalCount, documents), nil
}

func (r *UploadResolver) AuditLogs(ctx context.Context) (*[]LSIFUploadsAuditLogsResolver, error) {
	logs, err := r.uploadsSvc.GetAuditLogsForUpload(ctx, r.upload.ID)
	if err != nil {
		return nil, err
	}

	resolvers := make([]LSIFUploadsAuditLogsResolver, 0, len(logs))
	for _, log := range logs {
		resolvers = append(resolvers, NewLSIFUploadsAuditLogsResolver(log))
	}

	return &resolvers, nil
}
