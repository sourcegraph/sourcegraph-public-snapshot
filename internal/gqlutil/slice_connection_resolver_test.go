package gqlutil

import (
	"bytes"
	"context"
	"testing"
)

func TestSliceConnectionResolver(t *testing.T) {
	store := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
	}

	t.Run("paginated slice", func(t *testing.T) {
		ctx := context.Background()
		offset := 0
		limit := 3

		data := store[offset : offset+limit]
		resolver := NewSliceConnectionResolver(data, len(store), offset+limit)

		nodes := resolver.Nodes(ctx)
		if len(nodes) != limit {
			t.Fatalf("expected at most %d nodes, got %d", limit, len(nodes))
		}

		for idx, node := range nodes {
			expected := data[idx]
			if !bytes.Equal(node, expected) {
				t.Fatalf("expected node to be %s, got %s", string(expected), string(node))
			}
		}

		totalCount := resolver.TotalCount(ctx)
		if int(totalCount) != len(data) {
			t.Fatalf("expected totalCount to be %d, got %d", len(store), totalCount)
		}

		p := resolver.PageInfo(ctx)
		if p == nil {
			t.Fatal("expected pageInfo to be non-nil")
		}

		if p.hasNextPage != false {
			t.Fatalf("expected hasNextPage to be false, got %t", p.hasNextPage)
		}
	})

	t.Run("hasNextPage is true", func(t *testing.T) {
		ctx := context.Background()

		offset := 0
		limit := 2
		data := store[offset : offset+limit]

		resolver := NewSliceConnectionResolver(data, len(store), limit)
		p := resolver.PageInfo(ctx)
		if p == nil {
			t.Fatal("expected pageInfo to be non-nil")
		}

		nodes := resolver.Nodes(ctx)
		if len(nodes) > limit {
			t.Fatalf("expected at most %d nodes, got %d", limit, len(nodes))
		}

		for idx, node := range nodes {
			expected := data[idx]
			if !bytes.Equal(node, expected) {
				t.Fatalf("expected node to be %s, got %s", string(expected), string(node))
			}
		}

		totalCount := resolver.TotalCount(ctx)
		if int(totalCount) != len(store) {
			t.Fatalf("expected totalCount to be %d, got %d", len(store), totalCount)
		}

		if p.hasNextPage != true {
			t.Fatalf("expected hasNextPage to be true, got %t", p.hasNextPage)
		}
	})
}
