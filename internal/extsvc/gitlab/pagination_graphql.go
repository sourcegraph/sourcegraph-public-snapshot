package gitlab

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/extsvc/gitlab/graphql"
)

func newGraphQLPaginatedResult[T any](
	ctx context.Context, pageInfo *graphql.PageInfo, initial []T,
	nextPage func(ctx context.Context, cursor string) ([]T, *graphql.PageInfo, error),
) (*PaginatedResult[T], error) {
	var cursor string
	if pageInfo.HasNextPage {
		cursor = pageInfo.EndCursor
	}

	called := false
	return newPaginatedResult(func() ([]T, error) {
		if !called {
			called = true
			return initial, nil
		}

		if cursor == "" {
			return []T{}, nil
		}

		page, pageInfo, err := nextPage(ctx, cursor)
		if err != nil {
			return nil, err
		}

		if pageInfo.HasNextPage {
			cursor = pageInfo.EndCursor
		} else {
			cursor = ""
		}

		return page, nil
	})
}
