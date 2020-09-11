package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

type BatchInserter struct {
	db                dbutil.DB
	numColumns        int
	maxBatchSize      int
	batch             []interface{}
	queryPrefix       string
	queryPlaceholders []string
}

func NewBatchInserter(ctx context.Context, db dbutil.DB, tableName string, columnNames ...string) *BatchInserter {
	numColumns := len(columnNames)
	maxBatchSize := getMaxBatchSize(numColumns)
	queryPrefix := makeQueryPrefix(tableName, columnNames)
	queryPlaceholders := makeQueryPlaceholders(numColumns)

	return &BatchInserter{
		db:                db,
		numColumns:        numColumns,
		maxBatchSize:      maxBatchSize,
		batch:             make([]interface{}, 0, maxBatchSize),
		queryPrefix:       queryPrefix,
		queryPlaceholders: queryPlaceholders,
	}
}

func (i *BatchInserter) Insert(ctx context.Context, values ...interface{}) error {
	if len(values) != i.numColumns {
		return fmt.Errorf("expected %d values, got %d", i.numColumns, len(values))
	}

	i.batch = append(i.batch, values...)

	if len(i.batch) >= i.maxBatchSize {
		// Flush full batch
		return i.Flush(ctx)
	}

	return nil
}

// Flush ensures that all queued rows are inserted. This method must be invoked at the end
// of insertion to ensure that all records are flushed to the underlying Execable.
func (i *BatchInserter) Flush(ctx context.Context) error {
	if batch := i.pop(); len(batch) != 0 {
		// Create a query with enough placeholders to match the current batch size. This should
		// generally be the full queryPlaceholders slice, except for the last call to Flush which
		// may be a partial batch.
		query := i.queryPrefix + strings.Join(i.queryPlaceholders[:len(batch)/i.numColumns], ",")

		if _, err := i.db.ExecContext(ctx, query, batch...); err != nil {
			return err
		}
	}

	return nil
}

func (i *BatchInserter) pop() (batch []interface{}) {
	if len(i.batch) < i.maxBatchSize {
		batch, i.batch = i.batch, i.batch[:0]
		return batch
	}

	batch, i.batch = i.batch[:i.maxBatchSize], i.batch[i.maxBatchSize:]
	return batch
}

const maxNumPostgresParameters = 32767

func getMaxBatchSize(numColumns int) int {
	return (maxNumPostgresParameters / numColumns) * numColumns
}

func makeQueryPrefix(tableName string, columnNames []string) string {
	quotedColumnNames := make([]string, 0, len(columnNames))
	for _, columnName := range columnNames {
		quotedColumnNames = append(quotedColumnNames, fmt.Sprintf(`"%s"`, columnName))
	}

	return fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, tableName, strings.Join(quotedColumnNames, ","))
}

func makeQueryPlaceholders(numColumns int) []string {
	maxBatchSize := getMaxBatchSize(numColumns)

	queryPlaceholders := make([]string, 0, maxBatchSize/numColumns)
	for i := 0; i < maxNumPostgresParameters; i += numColumns {
		placeholders := make([]string, 0, numColumns)
		for j := 0; j < numColumns; j++ {
			placeholders = append(placeholders, fmt.Sprintf("$%d", i+j+1))
		}

		queryPlaceholders = append(queryPlaceholders, fmt.Sprintf("(%s)", strings.Join(placeholders, ",")))
	}

	return queryPlaceholders
}
