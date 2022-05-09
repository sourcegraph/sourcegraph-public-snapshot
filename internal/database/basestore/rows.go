package basestore

import (
	"database/sql"
	"time"
)

// CloseRows closes the given rows object. The resulting error is a multierror
// containing the error parameter along with any errors that occur during scanning
// or closing the rows object. The rows object is assumed to be non-nil.
//
// The signature of this function allows scan methods to be written uniformly:
//
//     func ScanThings(rows *sql.Rows, queryErr error) (_ []Thing, err error) {
//         if queryErr != nil {
//             return nil, queryErr
//         }
//         defer func() { err = CloseRows(rows, err) }()
//
//         // read things from rows
//     }
//
// Scan methods should be called directly with the results of `*store.Query` to
// ensure that the rows are always properly handled.
//
//     things, err := ScanThings(store.Query(ctx, query))
func CloseRows(rows *sql.Rows, err error) error {
	return combineErrors(err, rows.Close(), rows.Err())
}

// ScanSlice reads a slice of values from the given row object.
func ScanSlice[T any](rows *sql.Rows, queryErr error) (_ []T, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	var values []T
	for rows.Next() {
		var value T
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// ScanFirst reads a single value from the given row object.
func ScanFirst[T any](rows *sql.Rows, queryErr error) (value T, _ bool, err error) {
	if queryErr != nil {
		return value, false, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return value, false, err
		}

		return value, true, nil
	}

	return value, false, nil
}

// ScanStrings reads string values from the given row object.
var ScanStrings = ScanSlice[string]

// ScanFirstString reads string values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
var ScanFirstString = ScanFirst[string]

// ScanFirstNullString reads possibly null string values from the given row
// object and returns the first one. If no rows match the query, a false-valued
// flag is returned.
func ScanFirstNullString(rows *sql.Rows, queryErr error) (_ string, _ bool, err error) {
	value, ok, err := ScanFirst[sql.NullString](rows, queryErr)
	if err != nil || !ok {
		return "", false, err
	}

	return value.String, true, nil
}

// ScanInt is a convenience method to return an integer value and any query error from a given row object.
func ScanInt(row *sql.Row) (int, error) {
	var value int
	return value, row.Scan(&value)
}

// ScanInts reads integer values from the given row object.
var ScanInts = ScanSlice[int]

// ScanInt32s reads integer values from the given row object.
var ScanInt32s = ScanSlice[int32]

// ScanInt64s reads integer values from the given row object.
var ScanInt64s = ScanSlice[int64]

// ScanUint32s reads unsigned integer values from the given row object.
var Scanuint32s = ScanSlice[uint32]

// ScanFirstInt reads integer values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
var ScanFirstInt = ScanFirst[int]

// ScanFirstInt64 reads int64 values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
var ScanFirstInt64 = ScanFirst[int64]

// ScanFirstNullInt64 reads possibly null int64 values from the given row
// object and returns the first one. If no rows match the query, a false-valued
// flag is returned.
func ScanFirstNullInt64(rows *sql.Rows, queryErr error) (_ int64, _ bool, err error) {
	value, ok, err := ScanFirst[sql.NullInt64](rows, queryErr)
	if err != nil || !ok {
		return 0, false, err
	}

	return value.Int64, true, nil
}

// ScanFloats reads float values from the given row object.
var ScanFloats = ScanSlice[float64]

// ScanFirstFloat reads float values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
var ScanFirstFloat = ScanFirst[float64]

// ScanBools reads bool values from the given row object.
var ScanBools = ScanSlice[bool]

// ScanFirstBool reads bool values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
var ScanFirstBool = ScanFirst[bool]

// ScanTimes reads time values from the given row object.
var ScanTimes = ScanSlice[time.Time]

// ScanFirstTime reads time values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
var ScanFirstTime = ScanFirst[time.Time]
