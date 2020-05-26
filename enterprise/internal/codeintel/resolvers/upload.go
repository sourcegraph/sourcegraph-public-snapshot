package resolvers

import (
	"context"
	"encoding/base64"
	"strconv"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/graph-gophers/graphql-go/relay"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/lsifserver/client"
	"github.com/sourcegraph/sourcegraph/internal/lsif"
)

type lsifUploadResolver struct {
	repositoryResolver *graphqlbackend.RepositoryResolver
	lsifUpload         *lsif.LSIFUpload
}

var _ graphqlbackend.LSIFUploadResolver = &lsifUploadResolver{}

func (r *lsifUploadResolver) ID() graphql.ID {
	return marshalLSIFUploadGQLID(r.lsifUpload.ID)
}

func (r *lsifUploadResolver) ProjectRoot(ctx context.Context) (*graphqlbackend.GitTreeEntryResolver, error) {
	return resolvePath(ctx, r.lsifUpload.RepositoryID, r.lsifUpload.Commit, r.lsifUpload.Root)
}

func (r *lsifUploadResolver) InputCommit() string {
	return r.lsifUpload.Commit
}

func (r *lsifUploadResolver) InputRoot() string {
	return r.lsifUpload.Root
}

func (r *lsifUploadResolver) InputIndexer() string {
	return r.lsifUpload.Indexer
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

func (r *lsifUploadResolver) IsLatestForRepo() bool {
	return r.lsifUpload.VisibleAtTip
}

func (r *lsifUploadResolver) PlaceInQueue() *int32 {
	return r.lsifUpload.PlaceInQueue
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
	RepositoryID    graphql.ID
	Query           *string
	State           *string
	IsLatestForRepo *bool
	Limit           *int32
	NextURL         *string
}

type lsifUploadConnectionResolver struct {
	lsifserverClient *client.Client

	opt LSIFUploadsListOptions

	// cache results because they are used by multiple fields
	once               sync.Once
	uploads            []*lsif.LSIFUpload
	repositoryResolver *graphqlbackend.RepositoryResolver
	totalCount         *int
	nextURL            string
	err                error
}

var _ graphqlbackend.LSIFUploadConnectionResolver = &lsifUploadConnectionResolver{}

func (r *lsifUploadConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LSIFUploadResolver, error) {
	uploads, repositoryResolver, _, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LSIFUploadResolver
	for _, lsifUpload := range uploads {
		l = append(l, &lsifUploadResolver{
			repositoryResolver: repositoryResolver,
			lsifUpload:         lsifUpload,
		})
	}
	return l, nil
}

func (r *lsifUploadConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	_, _, count, _, err := r.compute(ctx)
	if count == nil || err != nil {
		return nil, err
	}

	c := int32(*count)
	return &c, nil
}

func (r *lsifUploadConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *lsifUploadConnectionResolver) compute(ctx context.Context) ([]*lsif.LSIFUpload, *graphqlbackend.RepositoryResolver, *int, string, error) {
	r.once.Do(func() {
		r.repositoryResolver, r.err = graphqlbackend.RepositoryByID(ctx, r.opt.RepositoryID)
		if r.err != nil {
			return
		}

		r.uploads, r.nextURL, r.totalCount, r.err = r.lsifserverClient.GetUploads(ctx, &struct {
			RepoID          api.RepoID
			Query           *string
			State           *string
			IsLatestForRepo *bool
			Limit           *int32
			Cursor          *string
		}{
			RepoID:          r.repositoryResolver.Type().ID,
			Query:           r.opt.Query,
			State:           r.opt.State,
			IsLatestForRepo: r.opt.IsLatestForRepo,
			Limit:           r.opt.Limit,
			Cursor:          r.opt.NextURL,
		})
	})

	return r.uploads, r.repositoryResolver, r.totalCount, r.nextURL, r.err
}

func marshalLSIFUploadGQLID(lsifUploadID int64) graphql.ID {
	return relay.MarshalID("LSIFUpload", lsifUploadID)
}

func unmarshalLSIFUploadGQLID(id graphql.ID) (lsifUploadID int64, err error) {
	// First, try to unmarshal the ID as a string and then convert it to an
	// integer. This is here to maintain backwards compatibility with the
	// src-cli lsif upload command, which constructs its own relay identifier
	// from a the string payload returned by the upload proxy.

	var lsifUploadIDString string
	err = relay.UnmarshalSpec(id, &lsifUploadIDString)
	if err == nil {
		lsifUploadID, err = strconv.ParseInt(lsifUploadIDString, 10, 64)
		return
	}

	// If it wasn't unmarshal-able as a string, it's a new-style int identifier
	err = relay.UnmarshalSpec(id, &lsifUploadID)
	return
}
