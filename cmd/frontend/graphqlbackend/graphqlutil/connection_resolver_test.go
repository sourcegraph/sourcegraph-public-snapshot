package graphqlutil

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

var TOTAL_COUNT = int32(10)

type TestConnectionNode struct {
	id int
}

func (n TestConnectionNode) ID() graphql.ID {
	return graphql.ID(fmt.Sprintf("%d", n.id))
}

type TestConnectionStore struct {
	t                  *testing.T
	TestComputeNodes   func(ctx context.Context, args *database.PaginationArgs)
	ComputeTotalCalled int
	ComputeNodesCalled int
}

func (s *TestConnectionStore) ComputeTotal(ctx context.Context) (*int32, error) {
	s.ComputeTotalCalled = s.ComputeTotalCalled + 1
	total := TOTAL_COUNT

	return &total, nil
}

func (s *TestConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*TestConnectionNode, error) {
	s.ComputeNodesCalled = s.ComputeNodesCalled + 1
	nodes := []*TestConnectionNode{{id: 0}, {id: 1}}

	s.TestComputeNodes(ctx, args)

	return nodes, nil
}

func (*TestConnectionStore) MarshalCursor(n *TestConnectionNode) (*string, error) {
	cursor := string(n.ID())

	return &cursor, nil
}

func (*TestConnectionStore) UnMarshalCursor(cursor string) (*int32, error) {
	num, err := strconv.Atoi(cursor)
	if err != nil {
		return nil, err
	}

	id := int32(num)

	return &id, nil
}

func TestConnectionTotalCount(t *testing.T) {
	ctx := context.Background()
	store := &TestConnectionStore{t, func(context.Context, *database.PaginationArgs) { return }, 0, 0}
	resolver := NewConnectionResolver[TestConnectionNode](store, nil)

	count, err := resolver.TotalCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(TOTAL_COUNT, count); diff != "" {
		t.Fatal(diff)
	}

	resolver.TotalCount(ctx)
	if diff := cmp.Diff(1, store.ComputeTotalCalled); diff != "" {
		t.Fatal(diff)
	}
}

func toInt32(n int) *int32 {
	num := int32(n)

	return &num
}

func TestConnectionNodesDefaultArgs(t *testing.T) {
	store := &TestConnectionStore{t,
		func(_ context.Context, args *database.PaginationArgs) {
			expectedArgs := &database.PaginationArgs{
				First:  toInt32(6),
				Last:   nil,
				After:  nil,
				Before: nil,
			}
			if diff := cmp.Diff(expectedArgs, args); diff != "" {
				t.Fatal(diff)
			}
		}, 0, 0}

	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{
		First:  toInt32(5),
		Last:   nil,
		After:  nil,
		Before: nil,
	},
	)

	ctx := context.Background()
	nodes, err := resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(2, len(nodes)); diff != "" {
		t.Fatal(diff)
	}

	resolver.Nodes(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionNodesLastArg(t *testing.T) {
	store := &TestConnectionStore{t,
		func(_ context.Context, args *database.PaginationArgs) {
			expectedArgs := &database.PaginationArgs{
				First:  nil,
				Last:   toInt32(6),
				After:  nil,
				Before: nil,
			}
			if diff := cmp.Diff(expectedArgs, args); diff != "" {
				t.Fatal(diff)
			}
		}, 0, 0}

	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{
		First:  nil,
		Last:   toInt32(5),
		After:  nil,
		Before: nil,
	},
	)

	ctx := context.Background()
	nodes, err := resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(2, len(nodes)); diff != "" {
		t.Fatal(diff)
	}

	resolver.Nodes(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionNodesAfterArgs(t *testing.T) {
	store := &TestConnectionStore{t,
		func(_ context.Context, args *database.PaginationArgs) {
			expectedArgs := &database.PaginationArgs{
				First:  toInt32(6),
				Last:   nil,
				After:  toInt32(0),
				Before: nil,
			}
			if diff := cmp.Diff(expectedArgs, args); diff != "" {
				t.Fatal(diff)
			}
		}, 0, 0}

	afterCursor := "0"
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{
		First:  toInt32(5),
		Last:   nil,
		After:  &afterCursor,
		Before: nil,
	},
	)

	ctx := context.Background()
	nodes, err := resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(2, len(nodes)); diff != "" {
		t.Fatal(diff)
	}

	resolver.Nodes(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionNodesBeforeArgs(t *testing.T) {
	store := &TestConnectionStore{t,
		func(_ context.Context, args *database.PaginationArgs) {
			expectedArgs := &database.PaginationArgs{
				First:  toInt32(6),
				Last:   nil,
				After:  nil,
				Before: toInt32(0),
			}
			if diff := cmp.Diff(expectedArgs, args); diff != "" {
				t.Fatal(diff)
			}
		}, 0, 0}

	beforeCursor := "0"
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{
		First:  toInt32(5),
		Last:   nil,
		After:  nil,
		Before: &beforeCursor,
	},
	)

	ctx := context.Background()
	nodes, err := resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(2, len(nodes)); diff != "" {
		t.Fatal(diff)
	}

	resolver.Nodes(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionNodesWithLimit(t *testing.T) {
	store := &TestConnectionStore{t,
		func(_ context.Context, args *database.PaginationArgs) {
			expectedArgs := &database.PaginationArgs{
				First:  toInt32(2),
				Last:   nil,
				After:  nil,
				Before: toInt32(0),
			}
			if diff := cmp.Diff(expectedArgs, args); diff != "" {
				t.Fatal(diff)
			}
		}, 0, 0}

	beforeCursor := "0"
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{
		First:  toInt32(1),
		Last:   nil,
		After:  nil,
		Before: &beforeCursor,
	},
	)

	ctx := context.Background()
	nodes, err := resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(1, len(nodes)); diff != "" {
		t.Fatal(diff)
	}

	resolver.Nodes(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfoDefault(t *testing.T) {
	ctx := context.Background()
	store := &TestConnectionStore{t, func(context.Context, *database.PaginationArgs) { return }, 0, 0}
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{First: nil, Last: nil, After: nil, Before: nil})

	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *startCursor); diff != "" {
		t.Fatal(diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("1", *endCursor); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(false, pageInfo.HasNextPage()); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(false, pageInfo.HasPreviousPage()); diff != "" {
		t.Fatal(diff)
	}

	resolver.PageInfo(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfoFirstPage(t *testing.T) {
	ctx := context.Background()
	store := &TestConnectionStore{t, func(context.Context, *database.PaginationArgs) { return }, 0, 0}
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{First: toInt32(1), Last: nil, After: nil, Before: nil})

	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *startCursor); diff != "" {
		t.Fatal(diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *endCursor); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(true, pageInfo.HasNextPage()); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(false, pageInfo.HasPreviousPage()); diff != "" {
		t.Fatal(diff)
	}

	resolver.PageInfo(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfoSecondPage(t *testing.T) {
	ctx := context.Background()
	store := &TestConnectionStore{t, func(context.Context, *database.PaginationArgs) { return }, 0, 0}
	afterCursor := "0"
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{First: toInt32(1), Last: nil, After: &afterCursor, Before: nil})

	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *startCursor); diff != "" {
		t.Fatal(diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *endCursor); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(true, pageInfo.HasNextPage()); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(true, pageInfo.HasPreviousPage()); diff != "" {
		t.Fatal(diff)
	}

	resolver.PageInfo(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfoBackwardFirstPage(t *testing.T) {
	ctx := context.Background()
	store := &TestConnectionStore{t, func(context.Context, *database.PaginationArgs) { return }, 0, 0}
	beforeCursor := "0"
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{First: nil, Last: toInt32(1), Before: &beforeCursor, After: nil})

	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *startCursor); diff != "" {
		t.Fatal(diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *endCursor); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(true, pageInfo.HasNextPage()); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(true, pageInfo.HasPreviousPage()); diff != "" {
		t.Fatal(diff)
	}

	resolver.PageInfo(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfoBackwardFirstPageWithoutCursor(t *testing.T) {
	ctx := context.Background()
	store := &TestConnectionStore{t, func(context.Context, *database.PaginationArgs) { return }, 0, 0}
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{First: nil, Last: toInt32(1), Before: nil, After: nil})

	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *startCursor); diff != "" {
		t.Fatal(diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *endCursor); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(false, pageInfo.HasNextPage()); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(true, pageInfo.HasPreviousPage()); diff != "" {
		t.Fatal(diff)
	}

	resolver.PageInfo(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfoBackwardLastPage(t *testing.T) {
	ctx := context.Background()
	store := &TestConnectionStore{t, func(context.Context, *database.PaginationArgs) { return }, 0, 0}
	afterCursor := "0"
	resolver := NewConnectionResolver[TestConnectionNode](store, &ConnectionResolverArgs{First: nil, Last: nil, Before: &afterCursor, After: nil})

	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("0", *startCursor); diff != "" {
		t.Fatal(diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff("1", *endCursor); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(true, pageInfo.HasNextPage()); diff != "" {
		t.Fatal(diff)
	}

	if diff := cmp.Diff(false, pageInfo.HasPreviousPage()); diff != "" {
		t.Fatal(diff)
	}

	resolver.PageInfo(ctx)
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}
