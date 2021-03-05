package batch

import (
	"context"
	"sync"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
	"github.com/sourcegraph/sourcegraph/internal/database/dbtesting"
)

func init() {
	dbtesting.DBNameSuffix = "batch"
}

func TestBatchInserter(t *testing.T) {
	if testing.Short() {
		t.Skip()
	}
	dbtesting.SetupGlobalTestDB(t)
	setupTestTable(t)

	expectedValues := makeTestValues(2, 0)
	testInsert(t, expectedValues)

	rows, err := dbconn.Global.Query("SELECT col1, col2, col3, col4, col5 from batch_inserter_test")
	if err != nil {
		t.Fatalf("unexpected error querying data: %s", err)
	}
	defer rows.Close()

	var values [][]interface{}
	for rows.Next() {
		var v1, v2, v3, v4 int
		var v5 string
		if err := rows.Scan(&v1, &v2, &v3, &v4, &v5); err != nil {
			t.Fatalf("unexpected error scanning data: %s", err)
		}

		values = append(values, []interface{}{v1, v2, v3, v4, v5})
	}

	if diff := cmp.Diff(expectedValues, values); diff != "" {
		t.Errorf("unexpected table contents (-want +got):\n%s", diff)
	}
}

func BenchmarkBatchInserter(b *testing.B) {
	dbtesting.SetupGlobalTestDB(b)
	setupTestTable(b)
	expectedValues := makeTestValues(10, 0)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testInsert(b, expectedValues)
	}
}

func BenchmarkBatchInserterLargePayload(b *testing.B) {
	dbtesting.SetupGlobalTestDB(b)
	setupTestTable(b)
	expectedValues := makeTestValues(10, 4096)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		testInsert(b, expectedValues)
	}
}

var setup sync.Once

func setupTestTable(t testing.TB) {
	setup.Do(func() {
		createTableQuery := `
			CREATE TABLE batch_inserter_test (
				id SERIAL,
				col1 integer NOT NULL,
				col2 integer NOT NULL,
				col3 integer NOT NULL,
				col4 integer NOT NULL,
				col5 text
			)
		`
		if _, err := dbconn.Global.Exec(createTableQuery); err != nil {
			t.Fatalf("unexpected error creating test table: %s", err)
		}
	})
}

func makeTestValues(tableSizeFactor, payloadSize int) [][]interface{} {
	var expectedValues [][]interface{}
	for i := 0; i < maxNumParameters*tableSizeFactor; i++ {
		expectedValues = append(expectedValues, []interface{}{
			i,
			i + 1,
			i + 2,
			i + 3,
			makePayload(payloadSize),
		})
	}

	return expectedValues
}

func makePayload(size int) string {
	s := make([]byte, 0, size)
	for i := 0; i < size; i++ {
		s = append(s, '!')
	}

	return string(s)
}

func testInsert(t testing.TB, expectedValues [][]interface{}) {
	ctx := context.Background()

	inserter := NewBatchInserter(ctx, dbconn.Global, "batch_inserter_test", "col1", "col2", "col3", "col4", "col5")
	for _, values := range expectedValues {
		if err := inserter.Insert(ctx, values...); err != nil {
			t.Fatalf("unexpected error inserting values: %s", err)
		}
	}

	if err := inserter.Flush(ctx); err != nil {
		t.Fatalf("unexpected error flushing: %s", err)
	}
}
