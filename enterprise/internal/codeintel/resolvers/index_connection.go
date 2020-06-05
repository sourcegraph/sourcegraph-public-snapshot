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

type lsifIndexConnectionResolver struct {
	db db.DB

	opt LSIFIndexesListOptions

	// cache results because they are used by multiple fields
	once               sync.Once
	indexes            []db.Index
	repositoryResolver *graphqlbackend.RepositoryResolver
	totalCount         *int
	nextURL            string
	err                error
}

var _ graphqlbackend.LSIFIndexConnectionResolver = &lsifIndexConnectionResolver{}

type LSIFIndexesListOptions struct {
	RepositoryID graphql.ID
	Query        *string
	State        *string
	Limit        *int32
	NextURL      *string
}

func (r *lsifIndexConnectionResolver) Nodes(ctx context.Context) ([]graphqlbackend.LSIFIndexResolver, error) {
	indexes, repositoryResolver, _, _, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	var l []graphqlbackend.LSIFIndexResolver
	for _, lsifIndex := range indexes {
		l = append(l, &lsifIndexResolver{
			repositoryResolver: repositoryResolver,
			lsifIndex:          lsifIndex,
		})
	}
	return l, nil
}

func (r *lsifIndexConnectionResolver) TotalCount(ctx context.Context) (*int32, error) {
	_, _, count, _, err := r.compute(ctx)
	if count == nil || err != nil {
		return nil, err
	}

	c := int32(*count)
	return &c, nil
}

func (r *lsifIndexConnectionResolver) PageInfo(ctx context.Context) (*graphqlutil.PageInfo, error) {
	_, _, _, nextURL, err := r.compute(ctx)
	if err != nil {
		return nil, err
	}

	if nextURL != "" {
		return graphqlutil.NextPageCursor(base64.StdEncoding.EncodeToString([]byte(nextURL))), nil
	}

	return graphqlutil.HasNextPage(false), nil
}

func (r *lsifIndexConnectionResolver) compute(ctx context.Context) ([]db.Index, *graphqlbackend.RepositoryResolver, *int, string, error) {
	r.once.Do(func() {
		var id int
		if r.opt.RepositoryID != "" {
			r.repositoryResolver, r.err = graphqlbackend.RepositoryByID(ctx, r.opt.RepositoryID)
			if r.err != nil {
				return
			}

			id = int(r.repositoryResolver.Type().ID)
		}

		query := ""
		if r.opt.Query != nil {
			query = *r.opt.Query
		}

		state := ""
		if r.opt.State != nil {
			state = strings.ToLower(*r.opt.State)
		}

		limit := DefaultIndexPageSize
		if r.opt.Limit != nil {
			limit = int(*r.opt.Limit)
		}

		offset := 0
		if r.opt.NextURL != nil {
			offset, _ = strconv.Atoi(*r.opt.NextURL)
		}

		indexes, totalCount, err := r.db.GetIndexes(ctx, db.GetIndexesOptions{
			RepositoryID: id,
			State:        state,
			Term:         query,
			Limit:        limit,
			Offset:       offset,
		})
		if err != nil {
			r.err = err
			return
		}

		cursor := ""
		if offset+len(indexes) < totalCount {
			cursor = fmt.Sprintf("%d", offset+len(indexes))
		}

		is := indexes

		r.indexes = is
		r.nextURL = cursor
		r.totalCount = &totalCount
	})

	return r.indexes, r.repositoryResolver, r.totalCount, r.nextURL, r.err
}
