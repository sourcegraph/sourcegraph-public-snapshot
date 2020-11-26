package graphql

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	gql "github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	store "github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/stores/dbstore"
	"github.com/sourcegraph/sourcegraph/internal/api"
)

type IndexResolver struct {
	index            store.Index
	locationResolver *CachedLocationResolver
}

func NewIndexResolver(index store.Index, locationResolver *CachedLocationResolver) gql.LSIFIndexResolver {
	return &IndexResolver{
		index:            index,
		locationResolver: locationResolver,
	}
}

func (r *IndexResolver) ID() graphql.ID            { return marshalLSIFIndexGQLID(int64(r.index.ID)) }
func (r *IndexResolver) InputCommit() string       { return r.index.Commit }
func (r *IndexResolver) QueuedAt() gql.DateTime    { return gql.DateTime{Time: r.index.QueuedAt} }
func (r *IndexResolver) State() string             { return strings.ToUpper(r.index.State) }
func (r *IndexResolver) Failure() *string          { return r.index.FailureMessage }
func (r *IndexResolver) StartedAt() *gql.DateTime  { return gql.DateTimeOrNil(r.index.StartedAt) }
func (r *IndexResolver) FinishedAt() *gql.DateTime { return gql.DateTimeOrNil(r.index.FinishedAt) }
func (r *IndexResolver) InputRoot() string         { return r.index.Root }
func (r *IndexResolver) Indexer() string           { return r.index.Indexer }
func (r *IndexResolver) IndexerArgs() []string     { return r.index.IndexerArgs }
func (r *IndexResolver) Outfile() *string          { return strPtr(r.index.Outfile) }
func (r *IndexResolver) PlaceInQueue() *int32      { return toInt32(r.index.Rank) }

func (r *IndexResolver) DockerSteps() []gql.DockerStepResolver {
	var steps []gql.DockerStepResolver
	for _, step := range r.index.DockerSteps {
		steps = append(steps, &dockerStepResolver{step})
	}

	return steps
}

func (r *IndexResolver) LogContents(ctx context.Context) (*string, error) {
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx); err != nil {
		if err == backend.ErrMustBeSiteAdmin {
			return nil, nil
		}

		return nil, err
	}

	// 🚨 SECURITY: Only site admins can view executor log contents.
	return strPtr(r.index.LogContents), nil
}

func (r *IndexResolver) ProjectRoot(ctx context.Context) (*gql.GitTreeEntryResolver, error) {
	return r.locationResolver.Path(ctx, api.RepoID(r.index.RepositoryID), r.index.Commit, r.index.Root)
}

type dockerStepResolver struct {
	step store.DockerStep
}

var _ gql.DockerStepResolver = &dockerStepResolver{}

func (r *dockerStepResolver) Root() string       { return r.step.Root }
func (r *dockerStepResolver) Image() string      { return r.step.Image }
func (r *dockerStepResolver) Commands() []string { return r.step.Commands }
