package graphqlutil

import (
	"context"
	"testing"
)

func TestSliceConnectionResolver(t *testing.T) {
	data := [][]byte{
		[]byte("a"),
		[]byte("b"),
		[]byte("c"),
		[]byte("d"),
		[]byte("e"),
		[]byte("f"),
		[]byte("g"),
		[]byte("h"),
		[]byte("i"),
		[]byte("j"),
		[]byte("k"),
		[]byte("l"),
	}

	ctx := context.Background()
	transformer := func(item []byte) (string, error) {
		return string(item), nil
	}

	resolver := NewSliceConnectionResolver(data, 2, 0, transformer)

	nodes, err := resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0] != "a" {
		t.Fatalf("expected first node to be a, got %s", nodes[0])
	}

	if nodes[1] != "b" {
		t.Fatalf("expected second node to be b, got %s", nodes[1])
	}

	totalCount := resolver.TotalCount(ctx)
	if int(totalCount) != len(data) {
		t.Fatalf("expected totalCount to be %d, got %d", len(data), totalCount)
	}

	resolver = NewSliceConnectionResolver(data, 2, 2, transformer)

	nodes, err = resolver.Nodes(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if len(nodes) != 2 {
		t.Fatalf("expected 2 nodes, got %d", len(nodes))
	}

	if nodes[0] != "c" {
		t.Fatalf("expected first node to be c, got %s", nodes[0])
	}

	if nodes[1] != "d" {
		t.Fatalf("expected second node to be d, got %s", nodes[1])
	}

	totalCount = resolver.TotalCount(ctx)
	if int(totalCount) != len(data) {
		t.Fatalf("expected totalCount to be %d, got %d", len(data), totalCount)
	}
}
