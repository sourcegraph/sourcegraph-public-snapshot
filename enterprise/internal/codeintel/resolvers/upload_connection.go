package resolvers

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"sync"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/graphqlbackend/graphqlutil"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/db"
)

type lsifUploadConnectionResolver struct {
	db db.DB

	opt LSIFUploadsListOptions

	// cache results because they are used by multiple fields
	once               sync.Once
	uploads            []db.Upload
	repositoryResolver *graphqlbackend.RepositoryResolver
	totalCount         *int
	nextURL            string
	err                error
}

var _ graphqlbackend.LSIFUploadConnectionResolver = &lsifUploadConnectionResolver{}

type LSIFUploadsListOptions struct {
	RepositoryID    graphql.ID
	Query           *string
	State           *string
	IsLatestForRepo *bool
	Limit           *int32
	NextURL         *string
}

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

func (r *lsifUploadConnectionResolver) compute(ctx context.Context) ([]db.Upload, *graphqlbackend.RepositoryResolver, *int, string, error) {
	r.once.Do(func() {
		r.repositoryResolver, r.err = graphqlbackend.RepositoryByID(ctx, r.opt.RepositoryID)
		if r.err != nil {
			return
		}

		id := int(r.repositoryResolver.Type().ID)
		query := ""
		if r.opt.Query != nil {
			query = *r.opt.Query
		}
		visibileAtTip := r.opt.IsLatestForRepo != nil && *r.opt.IsLatestForRepo

		state := ""
		if r.opt.State != nil {
			state = strings.ToLower(*r.opt.State)
		}

		limit := DefaultUploadPageSize
		if r.opt.Limit != nil {
			limit = int(*r.opt.Limit)
		}

		offset := 0
		if r.opt.NextURL != nil {
			offset, _ = strconv.Atoi(*r.opt.NextURL)
		}

		uploads, totalCount, err := r.db.GetUploadsByRepo(
			ctx,
			id,
			state,
			query,
			visibileAtTip,
			limit,
			offset,
		)
		if err != nil {
			r.err = err
			return
		}

		cursor := ""
		if offset+len(uploads) < totalCount {
			cursor = fmt.Sprintf("%d", offset+len(uploads))
		}

		us := uploads

		r.uploads = us
		r.nextURL = cursor
		r.totalCount = &totalCount
	})

	return r.uploads, r.repositoryResolver, r.totalCount, r.nextURL, r.err
}
