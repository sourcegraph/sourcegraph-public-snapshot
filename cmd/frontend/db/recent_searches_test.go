package db

import (
	"database/sql"
	"fmt"
	"math/rand"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func globalDB() *sql.DB { return dbconn.Global }

func TestRecentSearches_Add(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)
	q := fmt.Sprintf("fake query with random number %d", rand.Int())
	rs := &RecentSearches{globalDB}
	if err := rs.Log(ctx, q); err != nil {
		t.Fatal(err)
	}
	ss, err := rs.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(ss) != 1 {
		t.Fatalf("%d searches returned, want exactly 1", len(ss))
	}
	if ss[0] != q {
		t.Errorf("query is '%s', want '%s'", ss[0], q)
	}
}

func TestRecentSearches_DeleteExcessRows(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	rs := &RecentSearches{globalDB}
	t.Run("empty case", func(t *testing.T) {
		ctx := dbtesting.TestContext(t)
		if err := rs.Cleanup(ctx, 1); err != nil {
			t.Error(err)
		}
	})
	t.Run("single case", func(t *testing.T) {
		ctx := dbtesting.TestContext(t)
		q := "fake query"
		if err := rs.Log(ctx, q); err != nil {
			t.Fatal(err)
		}
		if err := rs.Cleanup(ctx, 2); err != nil {
			t.Error(err)
		}
		ss, err := rs.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(ss) != 1 {
			t.Errorf("recent_searches table has %d rows, want %d", len(ss), 1)
		}
	})
	t.Run("simple case", func(t *testing.T) {
		ctx := dbtesting.TestContext(t)
		limit := 10
		for i := 1; i <= limit+1; i++ {
			q := fmt.Sprintf("fake query for i = %d", i)
			if err := rs.Log(ctx, q); err != nil {
				t.Fatal(err)
			}
		}
		if err := rs.Cleanup(ctx, limit); err != nil {
			t.Fatal(err)
		}
		ss, err := rs.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(ss) != limit {
			t.Errorf("recent_searches table has %d rows, want %d", len(ss), limit)
		}
	})
	t.Run("id gap", func(t *testing.T) {
		ctx := dbtesting.TestContext(t)
		addQueryWithRandomId := func(q string) {
			insert := `INSERT INTO recent_searches (id, query) VALUES ((1e6*RANDOM())::int, $1)`
			if _, err := dbconn.Global.ExecContext(ctx, insert, q); err != nil {
				t.Fatalf("inserting '%s' into recent_searches table: %v", q, err)
			}
		}
		limit := 10
		for i := 1; i <= limit+1; i++ {
			q := fmt.Sprintf("fake query for i = %d", i)
			addQueryWithRandomId(q)
		}
		if err := rs.Cleanup(ctx, limit); err != nil {
			t.Fatal(err)
		}
		ss, err := rs.List(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if len(ss) != limit {
			t.Errorf("recent_searches table has %d rows, want %d", len(ss), limit)
		}
	})
}

func BenchmarkRecentSearches_AddAndDeleteExcessRows(b *testing.B) {
	rs := &RecentSearches{globalDB}
	ctx := dbtesting.TestContext(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		q := fmt.Sprintf("fake query for i = %d", i)
		if err := rs.Log(ctx, q); err != nil {
			b.Fatal(err)
		}
		if err := rs.Cleanup(ctx, b.N); err != nil {
			b.Fatal(err)
		}
	}
}
