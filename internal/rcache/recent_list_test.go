package rcache

import (
	"context"
	"fmt"
	"reflect"
	"testing"
)

func TestRecentListOK(t *testing.T) {
	SetupForTest(t)

	type testcase struct {
		key     string
		size    int
		inserts [][]byte
		want    [][]byte
	}

	cases := []testcase{
		{
			key:     "a",
			size:    3,
			inserts: bytes("a1", "a2", "a3"),
			want:    bytes("a3", "a2", "a1"),
		},
		{
			key:     "b",
			size:    3,
			inserts: bytes("a1", "a2", "a3", "a4", "a5", "a6"),
			want:    bytes("a6", "a5", "a4"),
		},
		{
			key:     "c",
			size:    3,
			inserts: bytes("a1", "a2"),
			want:    bytes("a2", "a1"),
		},
	}

	for _, c := range cases {
		r := NewRecentList(c.key, c.size)
		t.Run(fmt.Sprintf("size %d with %d entries", c.size, len(c.inserts)), func(t *testing.T) {
			for _, b := range c.inserts {
				if err := r.Insert(b); err != nil {
					t.Errorf("expected no error, got %q", err)
				}
			}
			got, err := r.All(context.Background())
			if err != nil {
				t.Errorf("expected no error, got %q", err)
			}
			s, err := r.Size()
			if err != nil {
				t.Errorf("expected no error, got %q", err)
			}
			if s != len(c.want) {
				t.Errorf("expected %d items, got %d instead", s, len(c.want))
			}
			if !reflect.DeepEqual(c.want, got) {
				t.Errorf("Expected %v, but got %v", _str(c.want...), _str(got...))
			}

		})
	}
}

func TestRecentListContextCancellation(t *testing.T) {
	r := NewRecentList("a", 3)
	err := r.Insert([]byte("a"))
	if err != nil {
		t.Errorf("expected no error, got %q", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = r.All(ctx)
	if err == nil {
		t.Fatal("expected error, got none")
	}
}

func _str(bs ...[]byte) []string {
	strs := make([]string, 0, len(bs))
	for _, b := range bs {
		strs = append(strs, string(b))
	}
	return strs
}
