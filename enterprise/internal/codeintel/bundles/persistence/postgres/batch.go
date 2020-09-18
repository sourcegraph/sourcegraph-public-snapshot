package postgres

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/db/dbutil"
)

// TODO - document
type BatchInserter struct {
	db           dbutil.DB
	numColumns   int
	maxBatchSize int
	batch        []interface{}
	queryPrefix  string
	querySuffix  string
}

// TODO - document
func NewBatchInserter(ctx context.Context, db dbutil.DB, tableName string, columnNames ...string) *BatchInserter {
	numColumns := len(columnNames)
	maxBatchSize := getMaxBatchSize(numColumns)
	queryPrefix := makeQueryPrefix(tableName, columnNames)
	querySuffix := makeQuerySuffix(numColumns)

	return &BatchInserter{
		db:           db,
		numColumns:   numColumns,
		maxBatchSize: maxBatchSize,
		batch:        make([]interface{}, 0, maxBatchSize),
		queryPrefix:  queryPrefix,
		querySuffix:  querySuffix,
	}
}

// TODO - document
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
		// ($xxxxx,$xxxxx,...)
		//  ^^^^^^^ * n - 1 (extra comma) + 2 (parens)
		rowSize := (7*i.numColumns + 1)

		// Determine number of rows being flushed
		numRows := len(batch) / i.numColumns

		// Account for commas separating rows
		n := numRows*rowSize + (numRows - 1)

		// Create a query with enough placeholders to match the current batch size. This should
		// generally be the full querySuffix string, except for the last call to Flush which
		// may be a partial batch.
		if _, err := i.db.ExecContext(ctx, i.queryPrefix+i.querySuffix[:n], batch...); err != nil {
			return err
		}
	}

	return nil
}

// TODO - document
func (i *BatchInserter) pop() (batch []interface{}) {
	if len(i.batch) < i.maxBatchSize {
		batch, i.batch = i.batch, i.batch[:0]
		return batch
	}

	batch, i.batch = i.batch[:i.maxBatchSize], i.batch[i.maxBatchSize:]
	return batch
}

// TODO - document
const maxNumPostgresParameters = 32767

// TODO - document
func getMaxBatchSize(numColumns int) int {
	return (maxNumPostgresParameters / numColumns) * numColumns
}

// TODO - document
func makeQueryPrefix(tableName string, columnNames []string) string {
	quotedColumnNames := make([]string, 0, len(columnNames))
	for _, columnName := range columnNames {
		quotedColumnNames = append(quotedColumnNames, fmt.Sprintf(`"%s"`, columnName))
	}

	return fmt.Sprintf(`INSERT INTO "%s" (%s) VALUES `, tableName, strings.Join(quotedColumnNames, ","))
}

// TODO - rename
// TODO - document
var m sync.Mutex
var querySuffixCache = map[int]string{}

// TODO - document
func makeQuerySuffix(numColumns int) string {
	m.Lock()
	defer m.Unlock()
	if cache, ok := querySuffixCache[numColumns]; ok {
		return cache
	}

	// TODO - clean up logic here
	qs := []byte{','}
	for i := 0; i < maxNumPostgresParameters; i++ {
		if i%numColumns == 0 {
			qs[len(qs)-1] = ')'
			qs = append(qs, ',', '(')
		}
		qs = append(qs, []byte(fmt.Sprintf("$%05d", i+1))...)
		qs = append(qs, ',')
	}
	qs[len(qs)-1] = ')'

	querySuffix := string(qs[2:])
	querySuffixCache[numColumns] = querySuffix
	return querySuffix
}
