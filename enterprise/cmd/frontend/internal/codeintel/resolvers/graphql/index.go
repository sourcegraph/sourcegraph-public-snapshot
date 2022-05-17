package graphql

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/opentracing/opentracing-go/log"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/enterprise/cmd/frontend/internal/codeintel/resolvers"
	"github.com/sourcegraph/sourcegraph/internal/api"
	policies "github.com/sourcegraph/sourcegraph/internal/codeintel/policies/enterprise"
	store "github.com/sourcegraph/sourcegraph/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/observation"
)

type IndexResolver struct {
	db               database.DB
	gitserver        policies.GitserverClient
	resolver         resolvers.Resolver
	index            store.Index
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
	traceErrs        *observation.ErrCollector
}

func NewIndexResolver(db database.DB, gitserver policies.GitserverClient, resolver resolvers.Resolver, index store.Index, prefetcher *Prefetcher, locationResolver *CachedLocationResolver, errTrace *observation.ErrCollector) gql.LSIFIndexResolver {
	if index.AssociatedUploadID != nil {
		// Request the next batch of upload fetches to contain the record's associated
		// upload id, if one exists it exists. This allows the prefetcher.GetUploadByID
		// invocation in the AssociatedUpload method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkUpload(*index.AssociatedUploadID)
	}

	return &IndexResolver{
		db:               db,
		gitserver:        gitserver,
		resolver:         resolver,
		index:            index,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
		traceErrs:        errTrace,
	}
}

func (r *IndexResolver) ID() graphql.ID            { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *IndexResolver) InputCommit() string       { return r.index.Commit }
func (r *IndexResolver) InputRoot() string         { return r.index.Root }
func (r *IndexResolver) InputIndexer() string      { return r.index.Indexer }
func (r *IndexResolver) QueuedAt() gql.DateTime    { return gql.DateTime{Time: r.index.QueuedAt} }
func (r *IndexResolver) Failure() *string          { return r.index.FailureMessage }
func (r *IndexResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.index.StartedAt) }
func (r *IndexResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.index.FinishedAt) }
func (r *IndexResolver) Steps() gql.IndexStepsResolver {
	return &indexStepsResolver{db: r.db, index: r.index}
}
func (r *IndexResolver) PlaceInQueue() *int32 { return toInt32(r.index.Rank) }

func (r *IndexResolver) State() string {
	state := strings.ToUpper(r.index.State)
	if state == "FAILED" {
		state = "ERRORED"
	}

	return state
}

func (r *IndexResolver) AssociatedUpload(ctx context.Context) (_ gql.LSIFUploadResolver, err error) {
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

	return NewUploadResolver(r.db, r.gitserver, r.resolver, upload, r.prefetcher, r.locationResolver, r.traceErrs), nil
}

func (r *IndexResolver) ProjectRoot(ctx context.Context) (_ *gql.GitTreeEntryResolver, err error) {
	defer r.traceErrs.Collect(&err, log.String("indexResolver.field", "projectRoot"))

	return r.locationResolver.Path(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, r.index.Root)
}

func (r *IndexResolver) Indexer() gql.CodeIntelIndexerResolver {
	// drop the tag if it exists
	if idx, ok := imageToIndexer[strings.Split(r.index.Indexer, ":")[0]]; ok {
		return idx
	}

	return &codeIntelIndexerResolver{name: r.index.Indexer}
}
