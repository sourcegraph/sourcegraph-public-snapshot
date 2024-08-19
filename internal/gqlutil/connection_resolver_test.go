package gqlutil

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/graph-gophers/graphql-go"

	"github.com/sourcegraph/sourcegraph/internal/database"
)

const testTotalCount = int32(10)

type testConnectionNode struct {
	id int
}

func (n testConnectionNode) ID() graphql.ID {
	return graphql.ID(fmt.Sprintf("%d", n.id))
}

type testConnectionStore struct {
	t                      *testing.T
	expectedPaginationArgs *database.PaginationArgs
	ComputeTotalCalled     int
	ComputeNodesCalled     int
}

func (s *testConnectionStore) testPaginationArgs(args *database.PaginationArgs) {
	if s.expectedPaginationArgs == nil {
		return
	}

	if diff := cmp.Diff(s.expectedPaginationArgs, args); diff != "" {
		s.t.Fatal(diff)
	}
}

func (s *testConnectionStore) ComputeTotal(ctx context.Context) (int32, error) {
	s.ComputeTotalCalled = s.ComputeTotalCalled + 1
	total := testTotalCount

	return int32(total), nil
}

func (s *testConnectionStore) ComputeNodes(ctx context.Context, args *database.PaginationArgs) ([]*testConnectionNode, error) {
	s.ComputeNodesCalled = s.ComputeNodesCalled + 1
	s.testPaginationArgs(args)

	nodes := []*testConnectionNode{{id: 0}, {id: 1}}

	return nodes, nil
}

func (*testConnectionStore) MarshalCursor(n *testConnectionNode, _ database.OrderBy) (*string, error) {
	cursor := string(n.ID())

	return &cursor, nil
}

func (*testConnectionStore) UnmarshalCursor(cursor string, _ database.OrderBy) ([]any, error) {
	i, err := strconv.Atoi(cursor)
	return []any{i}, err
}

func newInt32(n int) *int32 {
	num := int32(n)

	return &num
}

func withFirstCA(first int, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.First = newInt32(first)

	return a
}

func withLastCA(last int, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.Last = newInt32(last)

	return a
}

func withAfterCA(after string, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.After = &after

	return a
}

func withBeforeCA(before string, a *ConnectionResolverArgs) *ConnectionResolverArgs {
	a.Before = &before

	return a
}

func withFirstPA(first int, a *database.PaginationArgs) *database.PaginationArgs {
	a.First = &first

	return a
}

func withLastPA(last int, a *database.PaginationArgs) *database.PaginationArgs {
	a.Last = &last

	return a
}

func withAfterPA(after int, a *database.PaginationArgs) *database.PaginationArgs {
	a.After = []any{after}

	return a
}

func withBeforePA(before int, a *database.PaginationArgs) *database.PaginationArgs {
	a.Before = []any{before}

	return a
}

func TestConnectionTotalCount(t *testing.T) {
	ctx := context.Background()
	store := &testConnectionStore{t: t}
	resolver, err := NewConnectionResolver[*testConnectionNode](store, withFirstCA(1, &ConnectionResolverArgs{}), nil)
	if err != nil {
		t.Fatal(err)
	}

	count, err := resolver.TotalCount(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if count != testTotalCount {
		t.Fatalf("wrong total count. want=%d, have=%d", testTotalCount, count)
	}

	_, err = resolver.TotalCount(ctx)
	if err != nil {
		t.Fatalf("expected nil error when calling TotalCount, got %v", err)
	}
	if store.ComputeTotalCalled != 1 {
		t.Fatalf("wrong compute total called count. want=%d, have=%d", 1, store.ComputeTotalCalled)
	}
}

func testResolverNodesResponse(t *testing.T, resolver *ConnectionResolver[*testConnectionNode], store *testConnectionStore, count int, wantErr bool) {
	ctx := context.Background()
	nodes, err := resolver.Nodes(ctx)
	if wantErr {
		if err == nil {
			t.Fatalf("expected error, got %v", err)
		}
		return
	}
	if err != nil && !wantErr {
		t.Fatal(err)
	}

	if diff := cmp.Diff(count, len(nodes)); diff != "" {
		t.Fatal(diff)
	}

	_, err = resolver.Nodes(ctx)
	if err != nil {
		t.Fatalf("expected nil error when calling resolver.Nodes, got %v", err)
	}
	if store.ComputeNodesCalled != 1 {
		t.Fatalf("wrong compute nodes called count. want=%d, have=%d", 1, store.ComputeNodesCalled)
	}
}

func buildPaginationArgs() *database.PaginationArgs {
	args := database.PaginationArgs{
		OrderBy: database.OrderBy{{Field: "id"}},
	}

	return &args
}

func TestConnectionNodes(t *testing.T) {
	for _, test := range []struct {
		name           string
		connectionArgs *ConnectionResolverArgs
		options        *ConnectionResolverOptions

		wantError          bool
		wantPaginationArgs *database.PaginationArgs
		wantNodes          int
	}{
		{
			name:               "default",
			connectionArgs:     withFirstCA(5, &ConnectionResolverArgs{}),
			wantPaginationArgs: withFirstPA(6, buildPaginationArgs()),
			wantNodes:          2,
		},
		{
			name:               "last arg",
			wantPaginationArgs: withLastPA(6, buildPaginationArgs()),
			connectionArgs:     withLastCA(5, &ConnectionResolverArgs{}),
			wantNodes:          2,
		},
		{
			name:               "after arg",
			wantPaginationArgs: withAfterPA(0, withFirstPA(6, buildPaginationArgs())),
			connectionArgs:     withAfterCA("0", withFirstCA(5, &ConnectionResolverArgs{})),
			wantNodes:          2,
		},
		{
			name:               "before arg",
			wantPaginationArgs: withBeforePA(0, withLastPA(6, buildPaginationArgs())),
			connectionArgs:     withBeforeCA("0", withLastCA(5, &ConnectionResolverArgs{})),
			wantNodes:          2,
		},
		{
			name:               "with limit",
			wantPaginationArgs: withBeforePA(0, withLastPA(2, buildPaginationArgs())),
			connectionArgs:     withBeforeCA("0", withLastCA(1, &ConnectionResolverArgs{})),
			wantNodes:          1,
		},
		{
			name:           "no args supplied (skipArgValidation is false)",
			connectionArgs: &ConnectionResolverArgs{},
			options:        &ConnectionResolverOptions{AllowNoLimit: false},
			wantError:      true,
		},
		{
			name:           "no args supplied (skipArgValidation is true)",
			connectionArgs: &ConnectionResolverArgs{},
			options:        &ConnectionResolverOptions{AllowNoLimit: true},
			wantError:      false,
			wantNodes:      2,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &testConnectionStore{t: t, expectedPaginationArgs: test.wantPaginationArgs}
			resolver, err := NewConnectionResolver[*testConnectionNode](store, test.connectionArgs, test.options)
			if err != nil {
				t.Fatal(err)
			}

			testResolverNodesResponse(t, resolver, store, test.wantNodes, test.wantError)
		})
	}
}

type pageInfoResponse struct {
	startCursor     string
	endCursor       string
	hasNextPage     bool
	hasPreviousPage bool
}

func testResolverPageInfoResponse(t *testing.T, resolver *ConnectionResolver[*testConnectionNode], store *testConnectionStore, expectedResponse *pageInfoResponse) {
	ctx := context.Background()
	pageInfo, err := resolver.PageInfo(ctx)
	if err != nil {
		t.Fatal(err)
	}

	startCursor, err := pageInfo.StartCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(expectedResponse.startCursor, *startCursor); diff != "" {
		t.Fatal(diff)
	}

	endCursor, err := pageInfo.EndCursor()
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(expectedResponse.endCursor, *endCursor); diff != "" {
		t.Fatal(diff)
	}

	if expectedResponse.hasNextPage != pageInfo.HasNextPage() {
		t.Fatalf("hasNextPage should be %v, but is %v", expectedResponse.hasNextPage, pageInfo.HasNextPage())
	}
	if expectedResponse.hasPreviousPage != pageInfo.HasPreviousPage() {
		t.Fatalf("hasPreviousPage should be %v, but is %v", expectedResponse.hasPreviousPage, pageInfo.HasPreviousPage())
	}

	_, err = resolver.PageInfo(ctx)
	if err != nil {
		t.Fatalf("expected nil error when calling resolver.PageInfo, got %v", err)
	}
	if diff := cmp.Diff(1, store.ComputeNodesCalled); diff != "" {
		t.Fatal(diff)
	}
}

func TestConnectionPageInfo(t *testing.T) {
	for _, test := range []struct {
		name string
		args *ConnectionResolverArgs
		want *pageInfoResponse
	}{
		{
			name: "default",
			args: withFirstCA(20, &ConnectionResolverArgs{}),
			want: &pageInfoResponse{startCursor: "0", endCursor: "1", hasNextPage: false, hasPreviousPage: false},
		},
		{
			name: "first page",
			args: withFirstCA(1, &ConnectionResolverArgs{}),
			want: &pageInfoResponse{startCursor: "0", endCursor: "0", hasNextPage: true, hasPreviousPage: false},
		},
		{
			name: "second page",
			args: withAfterCA("0", withFirstCA(1, &ConnectionResolverArgs{})),
			want: &pageInfoResponse{startCursor: "0", endCursor: "0", hasNextPage: true, hasPreviousPage: true},
		},
		{
			name: "backward first page",
			args: withBeforeCA("0", withLastCA(1, &ConnectionResolverArgs{})),
			want: &pageInfoResponse{startCursor: "0", endCursor: "0", hasNextPage: true, hasPreviousPage: true},
		},
		{
			name: "backward first page without cursor",
			args: withLastCA(1, &ConnectionResolverArgs{}),
			want: &pageInfoResponse{startCursor: "0", endCursor: "0", hasNextPage: false, hasPreviousPage: true},
		},
		{
			name: "backward last page",
			args: withBeforeCA("0", withBeforeCA("0", withLastCA(20, &ConnectionResolverArgs{}))),
			want: &pageInfoResponse{startCursor: "1", endCursor: "0", hasNextPage: true, hasPreviousPage: false},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			store := &testConnectionStore{t: t}
			resolver, err := NewConnectionResolver[*testConnectionNode](store, test.args, nil)
			if err != nil {
				t.Fatal(err)
			}
			testResolverPageInfoResponse(t, resolver, store, test.want)
		})
	}
}
