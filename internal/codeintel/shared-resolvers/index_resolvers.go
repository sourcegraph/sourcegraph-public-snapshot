package shared

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/autoindexing/shared"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/types"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type LSIFIndexResolver interface {
	ID() graphql.ID
	InputCommit() string
	Tags(ctx context.Context) ([]string, error)
	InputRoot() string
	InputIndexer() string
	Indexer() types.CodeIntelIndexerResolver
	QueuedAt() DateTime
	State() string
	Failure() *string
	StartedAt() *DateTime
	FinishedAt() *DateTime
	Steps() IndexStepsResolver
	PlaceInQueue() *int32
	AssociatedUpload(ctx context.Context) (LSIFUploadResolver, error)
	ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error)
}

type indexResolver struct {
	svc              AutoIndexingService
	uploadsSvc       UploadsService
	policyResolver   PolicyResolver
	index            shared.Index
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewIndexResolver(svc AutoIndexingService, uploadsSvc UploadsService, policyResolver PolicyResolver, index shared.Index, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, errTrace *observation.ErrCollector) LSIFIndexResolver {
	if index.AssociatedUploadID != nil {
		// Request the next batch of upload fetches to contain the record's associated
		// upload id, if one exists it exists. This allows the prefetcher.GetUploadByID
		// invocation in the AssociatedUpload method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkUpload(*index.AssociatedUploadID)
	}

	return &indexResolver{
		svc:              svc,
		uploadsSvc:       uploadsSvc,
		policyResolver:   policyResolver,
		index:            index,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		traceErrs:        errTrace,
	}
}

func (r *indexResolver) ID() graphql.ID        { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *indexResolver) InputCommit() string   { return r.index.Commit }
func (r *indexResolver) InputRoot() string     { return r.index.Root }
func (r *indexResolver) InputIndexer() string  { return r.index.Indexer }
func (r *indexResolver) QueuedAt() DateTime    { return DateTime{Time: r.index.QueuedAt} }
func (r *indexResolver) Failure() *string      { return r.index.FailureMessage }
func (r *indexResolver) StartedAt() *DateTime  { return DateTimeOrNil(r.index.StartedAt) }
func (r *indexResolver) FinishedAt() *DateTime { return DateTimeOrNil(r.index.FinishedAt) }
func (r *indexResolver) Steps() IndexStepsResolver {
	return NewIndexStepsResolver(r.svc, r.index)
}
func (r *indexResolver) PlaceInQueue() *int32 { return toInt32(r.index.Rank) }

func (r *indexResolver) Tags(ctx context.Context) (tagsNames []string, err error) {
	tags, err := r.svc.GetListTags(ctx, api.RepoName(r.index.RepositoryName), r.index.Commit)
	if err != nil {
		return nil, err
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

func (r *indexResolver) AssociatedUpload(ctx context.Context) (_ LSIFUploadResolver, err error) {
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

	return NewUploadResolver(r.uploadsSvc, r.svc, r.policyResolver, upload, r.prefetcher, r.traceErrs), nil
}

func (r *indexResolver) ProjectRoot(ctx context.Context) (_ *gql.GitTreeEntryResolver, err error) {
	defer r.traceErrs.Collect(&err, log.String("indexResolver.field", "projectRoot"))

	return r.locationResolver.Path(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, r.index.Root)
}

func (r *indexResolver) Indexer() types.CodeIntelIndexerResolver {
	// drop the tag if it exists
	if idx, ok := types.ImageToIndexer[strings.Split(r.index.Indexer, ":")[0]]; ok {
		return idx
	}

	return types.NewCodeIntelIndexerResolver(r.index.Indexer)
}
