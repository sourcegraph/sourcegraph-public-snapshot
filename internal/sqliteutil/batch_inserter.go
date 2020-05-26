package sqliteutil

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// BatchInserter batches insertions to a single column in a SQLite database.
//
// The benchmark tests provided in this package show that 50% more rows can be
// inserted in the same time it takes for them to be inserted individually within
// a transaction.
//
// BenchmarkSQLiteInsertion-8                    	   40417	     29440 ns/op
// BenchmarkSQLiteInsertionInTransaction-8       	  214681	      5542 ns/op
// BenchmarkSQLiteInsertionWithBatchInserter-8   	  324998	      3701 ns/op
type BatchInserter struct {
	db                Execable
	numColumns        int
	maxBatchSize      int
	batch             []interface{}
	queryPrefix       string
	queryPlaceholders []string
}

// MaxNumSqliteParameters is the number of `?` placeholders that can be sent to SQLite without error.
const MaxNumSqliteParameters = 999

// Execable is the minimal common interface over sql.DB and sql.Tx required
// by BatchInserter.
type Execable interface {
	// ExecContext executes a query without returning any rows.
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// NewBatchInserter creates a new batch inserter.
func NewBatchInserter(db Execable, tableName string, columnNames ...string) *BatchInserter {
	numColumns := len(columnNames)
	maxBatchSize := (MaxNumSqliteParameters / numColumns) * numColumns

	placeholders := make([]string, numColumns)
	quotedColumnNames := make([]string, numColumns)
	for i, columnName := range columnNames {
		placeholders[i] = "?"
		quotedColumnNames[i] = fmt.Sprintf(`"%s"`, columnName)
	}

	queryPrefix := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, tableName, strings.Join(quotedColumnNames, ","))

	queryPlaceholders := make([]string, maxBatchSize/numColumns)
	for i := range queryPlaceholders {
		queryPlaceholders[i] = fmt.Sprintf("(%s)", strings.Join(placeholders, ","))
	}

	return &BatchInserter{
		db:                db,
		numColumns:        numColumns,
		maxBatchSize:      maxBatchSize,
		batch:             make([]interface{}, 0, maxBatchSize),
		queryPrefix:       queryPrefix,
		queryPlaceholders: queryPlaceholders,
	}
}

// Inserter enqueues the values of a single row for insertion. The given values must match up
// with the columnNames given at construction of the inserter.
func (bi *BatchInserter) Insert(ctx context.Context, values ...interface{}) error {
	if len(values) != bi.numColumns {
		return fmt.Errorf("expected %d values, got %d", bi.numColumns, len(values))
	}

	bi.batch = append(bi.batch, values...)

	if len(bi.batch) >= bi.maxBatchSize {
		// Flush full batch
		return bi.Flush(ctx)
	}

	return nil
}

// Flush ensures that all queued rows are inserted. This method must be invoked at the end
// of insertion to ensure that all records are flushed to the underlying Execable.
func (bi *BatchInserter) Flush(ctx context.Context) error {
	if batch := bi.pop(); len(batch) > 0 {
		// Create a query with enough placeholders to match the current batch size. This should
		// generally be the full queryPlaceholders slice, except for the last call to Flush which
		// may be a partial batch.
		query := bi.queryPrefix + strings.Join(bi.queryPlaceholders[:len(batch)/bi.numColumns], ",")

		if _, err := bi.db.ExecContext(ctx, query, batch...); err != nil {
			return err
		}
	}

	return nil
}

func (bi *BatchInserter) pop() (batch []interface{}) {
	if len(bi.batch) < bi.maxBatchSize {
		batch, bi.batch = bi.batch, bi.batch[:0]
		return batch
	}

	batch, bi.batch = bi.batch[:bi.maxBatchSize], bi.batch[bi.maxBatchSize:]
	return batch
}
