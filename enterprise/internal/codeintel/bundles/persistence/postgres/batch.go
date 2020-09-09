package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

type batchWriter struct {
	db                dbutil.DB
	numColumns        int
	maxBatchSize      int
	batch             []interface{}
	queryPrefix       string
	queryPlaceholders []string
}

const MaxNumSqliteParameters = 32767

func NewBatchInserter(ctx context.Context, db dbutil.DB, tableName string, columnNames ...string) (*batchWriter, error) {
	numColumns := len(columnNames)
	maxBatchSize := (MaxNumSqliteParameters / numColumns) * numColumns

	quotedColumnNames := make([]string, numColumns)
	for i, columnName := range columnNames {
		quotedColumnNames[i] = fmt.Sprintf(`"%s"`, columnName)
	}

	queryPrefix := fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, tableName, strings.Join(quotedColumnNames, ","))

	queryPlaceholders := make([]string, 0, maxBatchSize/numColumns)
	for i := 0; i < MaxNumSqliteParameters; i += numColumns {
		var placeholders []string
		for j := 0; j < numColumns; j++ {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+j+1))
		}

		queryPlaceholders = append(queryPlaceholders, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
	}

	return &batchWriter{
		db:                db,
		numColumns:        numColumns,
		maxBatchSize:      maxBatchSize,
		batch:             make([]interface{}, 0, maxBatchSize),
		queryPrefix:       queryPrefix,
		queryPlaceholders: queryPlaceholders,
	}, nil
}

func (w *batchWriter) Insert(ctx context.Context, values ...interface{}) error {
	if len(values) != w.numColumns {
		return fmt.Errorf("expected %d values, got %d", w.numColumns, len(values))
	}

	w.batch = append(w.batch, values...)

	if len(w.batch) >= w.maxBatchSize {
		// Flush full batch
		return w.Flush(ctx)
	}

	return nil
}

// Flush ensures that all queued rows are inserted. This method must be invoked at the end
// of insertion to ensure that all records are flushed to the underlying Execable.
func (w *batchWriter) Flush(ctx context.Context) error {
	if batch := w.pop(); len(batch) > 0 {
		// Create a query with enough placeholders to match the current batch size. This should
		// generally be the full queryPlaceholders slice, except for the last call to Flush which
		// may be a partial batch.
		query := w.queryPrefix + strings.Join(w.queryPlaceholders[:len(batch)/w.numColumns], ",")

		if _, err := w.db.ExecContext(ctx, query, batch...); err != nil {
			return err
		}
	}

	return nil
}

func (w *batchWriter) pop() (batch []interface{}) {
	if len(w.batch) < w.maxBatchSize {
		batch, w.batch = w.batch, w.batch[:0]
		return batch
	}

	batch, w.batch = w.batch[:w.maxBatchSize], w.batch[w.maxBatchSize:]
	return batch
}
