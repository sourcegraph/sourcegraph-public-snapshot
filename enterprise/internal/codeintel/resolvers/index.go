package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type lsifIndexResolver struct {
	repositoryResolver *graphqlbackend.RepositoryResolver
	lsifIndex          db.Index
}

var _ graphqlbackend.LSIFIndexResolver = &lsifIndexResolver{}

func (r *lsifIndexResolver) ID() graphql.ID      { return marshalLSIFIndexGQLID(int64(r.lsifIndex.ID)) }
func (r *lsifIndexResolver) InputCommit() string { return r.lsifIndex.Commit }
func (r *lsifIndexResolver) State() string       { return strings.ToUpper(r.lsifIndex.State) }

func (r *lsifIndexResolver) PlaceInQueue() *int32 {
	if r.lsifIndex.Rank == nil {
		return nil
	}

	v := int32(*r.lsifIndex.Rank)
	return &v
}

func (r *lsifIndexResolver) QueuedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.lsifIndex.QueuedAt}
}

func (r *lsifIndexResolver) StartedAt() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(r.lsifIndex.StartedAt)
}

func (r *lsifIndexResolver) FinishedAt() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(r.lsifIndex.FinishedAt)
}

func (r *lsifIndexResolver) ProjectRoot(ctx context.Context) (*graphqlbackend.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.lsifIndex.RepositoryID), r.lsifIndex.Commit, "")
}

func (r *lsifIndexResolver) Failure() graphqlbackend.LSIFIndexFailureReasonResolver {
	if r.lsifIndex.FailureSummary == nil {
		return nil
	}

	return &lsifIndexFailureReasonResolver{r.lsifIndex}
}
