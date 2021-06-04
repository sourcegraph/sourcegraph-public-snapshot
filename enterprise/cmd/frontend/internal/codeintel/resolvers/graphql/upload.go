package graphql

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"

	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type UploadResolver struct {
	upload           store.Upload
	prefetcher       *Prefetcher
	locationResolver *CachedLocationResolver
}

func NewUploadResolver(upload store.Upload, prefetcher *Prefetcher, locationResolver *CachedLocationResolver) gql.LSIFUploadResolver {
	if upload.AssociatedIndexID != nil {
		// Request the next batch of index fetches to contain the record's associated
		// index id, if one exists it exists. This allows the prefetcher.GetIndexByID
		// invocation in the AssociatedIndex method to batch its work with sibling
		// resolvers, which share the same prefetcher instance.
		prefetcher.MarkIndex(*upload.AssociatedIndexID)
	}

	return &UploadResolver{
		upload:           upload,
		prefetcher:       prefetcher,
		locationResolver: locationResolver,
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

func (r *UploadResolver) State() string {
	state := strings.ToUpper(r.upload.State)
	if state == "FAILED" {
		state = "ERRORED"
	}

	return state
}

func (r *UploadResolver) AssociatedIndex(ctx context.Context) (gql.LSIFIndexResolver, error) {
	// TODO - why are a bunch of them zero?
	if r.upload.AssociatedIndexID == nil || *r.upload.AssociatedIndexID == 0 {
		return nil, nil
	}

	index, exists, err := r.prefetcher.GetIndexByID(ctx, *r.upload.AssociatedIndexID)
	if err != nil || !exists {
		return nil, err
	}

	return NewIndexResolver(index, r.prefetcher, r.locationResolver), nil
}

func (r *UploadResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return r.locationResolver.Path(ctx, api.RepoID(r.upload.RepositoryID), r.upload.Commit, r.upload.Root)
}
