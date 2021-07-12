package graphql

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type IndexResolver struct {
	index            store.Index
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
}

func NewIndexResolver(index store.Index, prefetcher *Prefetcher, locationResolver *CachedLocationResolver) gql.LSIFIndexResolver {
	if index.AssociatedUploadID != nil {
		// Request the next batch of upload fetches to contain the record's associated
		// upload id, if one exists it exists. This allows the prefetcher.GetUploadByID
		// invocation in the AssociatedUpload method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkUpload(*index.AssociatedUploadID)
	}

	return &IndexResolver{
		index:            index,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
	}
}

func (r *IndexResolver) ID() graphql.ID                { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *IndexResolver) InputCommit() string           { return r.index.Commit }
func (r *IndexResolver) InputRoot() string             { return r.index.Root }
func (r *IndexResolver) InputIndexer() string          { return r.index.Indexer }
func (r *IndexResolver) QueuedAt() gql.DateTime        { return gql.DateTime{Time: r.index.QueuedAt} }
func (r *IndexResolver) Failure() *string              { return r.index.FailureMessage }
func (r *IndexResolver) StartedAt() *gql.DateTime      { return gql.DateTimeOrNil(r.index.StartedAt) }
func (r *IndexResolver) FinishedAt() *gql.DateTime     { return gql.DateTimeOrNil(r.index.FinishedAt) }
func (r *IndexResolver) Steps() gql.IndexStepsResolver { return &indexStepsResolver{index: r.index} }
func (r *IndexResolver) PlaceInQueue() *int32          { return toInt32(r.index.Rank) }

func (r *IndexResolver) State() string {
	state := strings.ToUpper(r.index.State)
	if state == "FAILED" {
		state = "ERRORED"
	}

	return state
}

func (r *IndexResolver) AssociatedUpload(ctx context.Context) (gql.LSIFUploadResolver, error) {
	if r.index.AssociatedUploadID == nil {
		return nil, nil
	}

	upload, exists, err := r.prefetcher.GetUploadByID(ctx, *r.index.AssociatedUploadID)
	if err != nil || !exists {
		return nil, err
	}

	return NewUploadResolver(upload, r.prefetcher, r.locationResolver), nil
}

func (r *IndexResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return r.locationResolver.Path(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, r.index.Root)
}
