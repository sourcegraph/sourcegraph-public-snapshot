package db

import (
	"database/sql"
	"fmt"
	"math/rand"
	"reflect"
	"testing"

	"github.com/sourcegraph/sourcegraph/pkg/db/dbconn"
	"github.com/sourcegraph/sourcegraph/pkg/db/dbtesting"
)

func globalDB() *sql.DB { return dbconn.Global }

func TestRecentSearches_Log(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	ctx := dbtesting.TestContext(t)
	q := fmt.Sprintf("fake query with random number %d", rand.Int())
	rs := &RecentSearches{}
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

func TestRecentSearches_Cleanup(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	rs := &RecentSearches{}
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

func BenchmarkRecentSearches_LogAndCleanup(b *testing.B) {
	rs := &RecentSearches{}
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

func TestRecentSearches_Top(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	tests := []struct {
		name    string
		queries []string
		n       int32
		want    map[string]int32
		wantErr bool
	}{
		{name: "empty case", queries: nil, n: 10, want: map[string]int32{}, wantErr: false},
		{name: "a", queries: []string{"a"}, n: 10, want: map[string]int32{"a": 1}, wantErr: false},
		{name: "a a", queries: []string{"a", "a"}, n: 10, want: map[string]int32{"a": 2}, wantErr: false},
		{name: "b a", queries: []string{"b", "a"}, n: 10, want: map[string]int32{"a": 1, "b": 1}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rs := &RecentSearches{}
			ctx := dbtesting.TestContext(t)
			for _, q := range tt.queries {
				if err := rs.Log(ctx, q); err != nil {
					t.Fatal(err)
				}
			}
			got, err := rs.Top(ctx, tt.n)
			if (err != nil) != tt.wantErr {
				t.Errorf("RecentSearches.Top() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RecentSearches.Top() = %v, want %v", got, tt.want)
			}
		})
	}
}
