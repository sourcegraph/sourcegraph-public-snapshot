package resolvers

import (
	"context"
	"encoding/base64"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/enterprise/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type lsifUploadResolver struct {
	lsifUpload *lsif.LSIFUpload
}

var _ graphqlbackend.LSIFUploadResolver = &lsifUploadResolver{}

func (r *lsifUploadResolver) ID() graphql.ID {
	return marshalLSIFUploadGQLID(r.lsifUpload.ID)
}

func (r *lsifUploadResolver) ProjectRoot(ctx context.Context) (*graphqlbackend.GitTreeEntryResolver, error) {
	repo, err := backend.Repos.GetByName(ctx, api.RepoName(r.lsifUpload.Repository))
	if err != nil {
		return nil, err
	}

	repoResolver := graphqlbackend.NewRepositoryResolver(repo)
	commitResolver, err := repoResolver.Commit(ctx, &graphqlbackend.RepositoryCommitArgs{Rev: r.lsifUpload.Commit})
	if err != nil {
		return nil, err
	}

	return graphqlbackend.NewGitTreeEntryResolver(commitResolver, graphqlbackend.CreateFileInfo(r.lsifUpload.Root, true)), nil
}

func (r *lsifUploadResolver) State() string {
	return strings.ToUpper(r.lsifUpload.State)
}

func (r *lsifUploadResolver) Failure() graphqlbackend.LSIFUploadFailureReasonResolver {
	if r.lsifUpload.FailureSummary == nil {
		return nil
	}

	return &lsifUploadFailureReasonResolver{r.lsifUpload}
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

type lsifUploadFailureReasonResolver struct {
	lsifUpload *lsif.LSIFUpload
}

var _ graphqlbackend.LSIFUploadFailureReasonResolver = &lsifUploadFailureReasonResolver{}

func (r *lsifUploadFailureReasonResolver) Summary() string {
	if r.lsifUpload.FailureSummary == nil {
		return ""
	}

	return *r.lsifUpload.FailureSummary
}

func (r *lsifUploadFailureReasonResolver) Stacktrace() string {
	if r.lsifUpload.FailureStacktrace == nil {
		return ""
	}

	return *r.lsifUpload.FailureStacktrace
}

type LSIFUploadsListOptions struct {
	State   string
	Query   *string
	Limit   *int32
	NextURL *string
}

type lsifUploadConnectionResolver struct {
	opt LSIFUploadsListOptions

	// cache results because they are used by multiple fields
	once       sync.Once
	uploads    []*lsif.LSIFUpload
	totalCount *int
	nextURL    string
	err        error
}

var _ graphqlbackend.LSIFUploadConnectionResolver = &lsifUploadConnectionResolver{}

func (r *lsifUploadConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LSIFUploadResolver, error) {
	uploads, _, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LSIFUploadResolver
	for _, lsifUpload := range uploads {
		l = append(l, &lsifUploadResolver{
			lsifUpload: lsifUpload,
		})
	}
	return l, nil
}

func (r *lsifUploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	_, count, _, err := r.compute(ctx)
	if count == nil || err != nil {
		return nil, err
	}

	c := int32(*count)
	return &c, nil
}

func (r *lsifUploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *lsifUploadConnectionResolver) compute(ctx context.Context) ([]*lsif.LSIFUpload, *int, string, error) {
	r.once.Do(func() {
		r.uploads, r.nextURL, r.totalCount, r.err = client.DefaultClient.GetUploads(ctx, &struct {
			State  string
			Query  *string
			Limit  *int32
			Cursor *string
		}{
			State:  r.opt.State,
			Query:  r.opt.Query,
			Limit:  r.opt.Limit,
			Cursor: r.opt.NextURL,
		})
	})

	return r.uploads, r.totalCount, r.nextURL, r.err
}

type lsifUploadStatsResolver struct {
	stats *lsif.LSIFUploadStats
}

var _ graphqlbackend.LSIFUploadStatsResolver = &lsifUploadStatsResolver{}

func (r *lsifUploadStatsResolver) ID() graphql.ID {
	return marshalLSIFUploadStatsGQLID(lsifUploadStatsGQLID)
}

func (r *lsifUploadStatsResolver) ProcessingCount() int32 { return r.stats.ProcessingCount }
func (r *lsifUploadStatsResolver) ErroredCount() int32    { return r.stats.ErroredCount }
func (r *lsifUploadStatsResolver) CompletedCount() int32  { return r.stats.CompletedCount }
func (r *lsifUploadStatsResolver) QueuedCount() int32     { return r.stats.QueuedCount }

func marshalLSIFUploadGQLID(lsifUploadID string) graphql.ID {
	return relay.MarshalID("LSIFUpload", lsifUploadID)
}

func unmarshalLSIFUploadGQLID(id graphql.ID) (lsifUploadID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifUploadID)
	return
}

func marshalLSIFUploadStatsGQLID(lsifUploadStatsID string) graphql.ID {
	return relay.MarshalID("LSIFUploadStats", lsifUploadStatsID)
}

func unmarshalLSIFUploadStatsGQLID(id graphql.ID) (lsifUploadStatsID string, err error) {
	err = relay.UnmarshalSpec(id, &lsifUploadStatsID)
	return
}
