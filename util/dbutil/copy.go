package dbutil

import (
	"database/sql"
	"fmt"
	"reflect"

	"github.com/lib/pq"
	"github.com/sqs/modl"
)

type Preparer interface {
	Prepare(sql string) (*sql.Stmt, error)
}

// GetPreparer extracts a Preparer from a *modl.Transaction or a
// ThreadSafeTx.
func GetPreparer(v interface{}) Preparer {
	switch v := v.(type) {
	case Preparer:
		return v
	case *modl.Transaction:
		return v.Tx
	default:
		panic(fmt.Sprintf("GetPreparer: can't extract Preparer from %T", v))
	}
}

// Copy executes a PostgreSQL COPY statement. All of the rows must have the same
// data type, which must be mapped to a table using modl.
//
// NOTE: It's probably not safe to call Copy on a database handle
// that's not a transaction (for database/sql concurrency/reuse
// reason).
func Copy(tx Preparer, tableName string, rows interface{}, omitCols ...int) error {
	rowsType := reflect.TypeOf(rows)
	if k := rowsType.Kind(); k != reflect.Slice && k != reflect.Array {
		panic("rows must be a slice or array")
	}
	rowType := rowsType.Elem()

	colNames := ColumnNames(rowType)
	if len(omitCols) > 0 {
		for _, c := range omitCols {
			colNames = append(colNames[:c], colNames[c+1:]...)
		}
	}

	stmt, err := tx.Prepare(pq.CopyIn(tableName, colNames...))
	if err != nil {
		return err
	}

	rowsVal := reflect.ValueOf(rows)
	for i := 0; i < rowsVal.Len(); i++ {
		rowVal := rowsVal.Index(i)
		colVals := ColumnValues(rowVal)
		if len(omitCols) > 0 {
			for _, c := range omitCols {
				colVals = append(colVals[:c], colVals[c+1:]...)
			}
		}
		_, err := stmt.Exec(colVals...)
		if err != nil {
			return err
		}
	}

	_, err = stmt.Exec()
	if err != nil {
		return err
	}

	err = stmt.Close()
	if err != nil {
		return err
	}

	return nil
}
