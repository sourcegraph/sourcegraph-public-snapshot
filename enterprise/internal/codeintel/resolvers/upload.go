package resolvers

import (
	"context"
	"strings"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type lsifUploadResolver struct {
	repositoryResolver *graphqlbackend.RepositoryResolver
	lsifUpload         db.Upload
}

var _ graphqlbackend.LSIFUploadResolver = &lsifUploadResolver{}

func (r *lsifUploadResolver) ID() graphql.ID        { return marshalLSIFUploadGQLID(int64(r.lsifUpload.ID)) }
func (r *lsifUploadResolver) InputCommit() string   { return r.lsifUpload.Commit }
func (r *lsifUploadResolver) InputRoot() string     { return r.lsifUpload.Root }
func (r *lsifUploadResolver) InputIndexer() string  { return r.lsifUpload.Indexer }
func (r *lsifUploadResolver) State() string         { return strings.ToUpper(r.lsifUpload.State) }
func (r *lsifUploadResolver) IsLatestForRepo() bool { return r.lsifUpload.VisibleAtTip }

func (r *lsifUploadResolver) PlaceInQueue() *int32 {
	if r.lsifUpload.Rank == nil {
		return nil
	}

	v := int32(*r.lsifUpload.Rank)
	return &v
}

func (r *lsifUploadResolver) UploadedAt() graphqlbackend.DateTime {
	return graphqlbackend.DateTime{Time: r.lsifUpload.UploadedAt}
}

func (r *lsifUploadResolver) StartedAt() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(r.lsifUpload.StartedAt)
}

func (r *lsifUploadResolver) FinishedAt() *graphqlbackend.DateTime {
	return graphqlbackend.DateTimeOrNil(r.lsifUpload.FinishedAt)
}

func (r *lsifUploadResolver) ProjectRoot(ctx context.Context) (*graphqlbackend.GitTreeEntryResolver, error) {
	return resolvePath(ctx, api.RepoID(r.lsifUpload.RepositoryID), r.lsifUpload.Commit, r.lsifUpload.Root)
}

func (r *lsifUploadResolver) Failure() graphqlbackend.LSIFUploadFailureReasonResolver {
	if r.lsifUpload.FailureSummary == nil {
		return nil
	}

	return &lsifUploadFailureReasonResolver{r.lsifUpload}
}
