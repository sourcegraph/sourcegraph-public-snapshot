package db

import (
	"fmt"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
	"math/rand"
	"testing"
)

func TestSearches_Add(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)

	t.Run("basic usage", func(t *testing.T) {
		q := fmt.Sprintf("fake query with random number %d", rand.Int())
		if err := Searches.Add(ctx, q, 100); err != nil {
			t.Fatal(err)
		}
		ss, err := Searches.Get(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(ss) != 1 {
			t.Fatalf("%d searches returned, want exactly 1", len(ss))
		}
		if ss[0] != q {
			t.Errorf("query is '%s', want '%s'", ss[0], q)
		}
	})
	t.Run("row count limit", func(t *testing.T) {
		limit := 10
		for i := 1; i <= limit * 2; i++ {
			q := fmt.Sprintf("fake query for i = %d", i)
			if err := Searches.Add(ctx, q, limit); err != nil {
				t.Fatal(err)
			}
			ss, err := Searches.Get(ctx)
			if err != nil {
				t.Fatal(err)
			}
			if i > limit {
				if len(ss) != limit {
					t.Errorf("for i = %d, got %d searches, want %d", i, len(ss), limit)
				}
			}
		}
	})
}

func BenchmarkSearches_Add(b *testing.B) {
	ctx := dbtesting.TestContext(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := fmt.Sprintf("fake query for i = %d", i)
		if err := Searches.Add(ctx, q, b.N); err != nil {
			b.Fatal(err)
		}
	}
}
