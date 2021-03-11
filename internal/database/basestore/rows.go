package basestore

import (
	"database/sql"
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

// ScanStrings reads string values from the given row object.
func ScanStrings(rows *sql.Rows, queryErr error) (_ []string, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	var values []string
	for rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// ScanFirstString reads string values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
func ScanFirstString(rows *sql.Rows, queryErr error) (_ string, _ bool, err error) {
	if queryErr != nil {
		return "", false, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	if rows.Next() {
		var value string
		if err := rows.Scan(&value); err != nil {
			return "", false, err
		}

		return value, true, nil
	}

	return "", false, nil
}

// ScanFirstNullString reads possibly null string values from the given row
// object and returns the first one. If no rows match the query, a false-valued
// flag is returned.
func ScanFirstNullString(rows *sql.Rows, queryErr error) (_ string, _ bool, err error) {
	if queryErr != nil {
		return "", false, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	if rows.Next() {
		var value sql.NullString
		if err := rows.Scan(&value); err != nil {
			return "", false, err
		}

		return value.String, true, nil
	}

	return "", false, nil
}

// ScanInts reads integer values from the given row object.
func ScanInts(rows *sql.Rows, queryErr error) (_ []int, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	var values []int
	for rows.Next() {
		var value int
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// ScanInt32s reads integer values from the given row object.
func ScanInt32s(rows *sql.Rows, queryErr error) (_ []int32, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	var values []int32
	for rows.Next() {
		var value int32
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// ScanFirstInt reads integer values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
func ScanFirstInt(rows *sql.Rows, queryErr error) (_ int, _ bool, err error) {
	if queryErr != nil {
		return 0, false, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	if rows.Next() {
		var value int
		if err := rows.Scan(&value); err != nil {
			return 0, false, err
		}

		return value, true, nil
	}

	return 0, false, nil
}

// ScanFloats reads float values from the given row object.
func ScanFloats(rows *sql.Rows, queryErr error) (_ []float64, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	var values []float64
	for rows.Next() {
		var value float64
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// ScanFirstFloat reads float values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
func ScanFirstFloat(rows *sql.Rows, queryErr error) (_ float64, _ bool, err error) {
	if queryErr != nil {
		return 0, false, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	if rows.Next() {
		var value float64
		if err := rows.Scan(&value); err != nil {
			return 0, false, err
		}

		return value, true, nil
	}

	return 0, false, nil
}

// ScanBools reads bool values from the given row object.
func ScanBools(rows *sql.Rows, queryErr error) (_ []bool, err error) {
	if queryErr != nil {
		return nil, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	var values []bool
	for rows.Next() {
		var value bool
		if err := rows.Scan(&value); err != nil {
			return nil, err
		}

		values = append(values, value)
	}

	return values, nil
}

// ScanFirstBool reads bool values from the given row object and returns the first one.
// If no rows match the query, a false-valued flag is returned.
func ScanFirstBool(rows *sql.Rows, queryErr error) (value bool, exists bool, err error) {
	if queryErr != nil {
		return false, false, queryErr
	}
	defer func() { err = CloseRows(rows, err) }()

	if rows.Next() {
		if err := rows.Scan(&value); err != nil {
			return false, false, err
		}

		return value, true, nil
	}

	return false, false, nil
}
