package basestore

import (
	"database/sql"
	"time"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
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

// NewSliceScanner returns a basestore scanner function that returns all
// the values of a query result. The given function is invoked multiple
// times with a SQL rows object to scan a single value.
func NewSliceScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows *sql.Rows, queryErr error) ([]T, error) {
	return func(rows *sql.Rows, queryErr error) (values []T, err error) {
		if queryErr != nil {
			return nil, queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		for rows.Next() {
			value, err := f(rows)
			if err != nil {
				return nil, err
			}

			values = append(values, value)
		}

		return values, nil
	}
}

// NewSliceWithCountScanner returns a basestore scanner function that returns all
// the values of the query result, as well as the count from the `COUNT(*) OVER()`
// window function. The given function is invoked multiple times with a SQL rows
// object to scan a single value.
// Example query that would avail of this function, where we want only 10 rows but still
// the count of everything that would have been returned, without performing two separate queries:
// SELECT u.id, COUNT(*) OVER() as count FROM users LIMIT 10
func NewSliceWithCountScanner[T any](f func(dbutil.Scanner) (T, int, error)) func(rows *sql.Rows, queryErr error) ([]T, int, error) {
	return func(rows *sql.Rows, queryErr error) (values []T, totalCount int, err error) {
		if queryErr != nil {
			return nil, 0, queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		for rows.Next() {
			value, count, err := f(rows)
			if err != nil {
				return nil, 0, err
			}

			totalCount = count
			values = append(values, value)
		}

		return values, totalCount, nil
	}
}

// NewFirstScanner returns a basestore scanner function that returns the
// first value of a query result (assuming there is at most one value).
// The given function is invoked with a SQL rows object to scan a single
// value.
func NewFirstScanner[T any](f func(dbutil.Scanner) (T, error)) func(rows *sql.Rows, queryErr error) (T, bool, error) {
	return func(rows *sql.Rows, queryErr error) (value T, _ bool, err error) {
		if queryErr != nil {
			return value, false, queryErr
		}
		defer func() { err = CloseRows(rows, err) }()

		if !rows.Next() {
			return value, false, nil
		}

		value, err = f(rows)
		return value, true, err
	}
}

// ScanAny scans a single T value from the given scanner.
func ScanAny[T any](s dbutil.Scanner) (value T, err error) {
	err = s.Scan(&value)
	return
}

// ScanNullString scans a single nullable string from the given scanner.
func ScanNullString(s dbutil.Scanner) (string, error) {
	var value sql.NullString
	if err := s.Scan(&value); err != nil {
		return "", err
	}

	return value.String, nil
}

// ScanNullInt64 scans a single int64 from the given scanner.
func ScanNullInt64(s dbutil.Scanner) (int64, error) {
	var value sql.NullInt64
	if err := s.Scan(&value); err != nil {
		return 0, err
	}

	return value.Int64, nil
}

var (
	ScanInt             = ScanAny[int]
	ScanStrings         = NewSliceScanner(ScanAny[string])
	ScanFirstString     = NewFirstScanner(ScanAny[string])
	ScanFirstNullString = NewFirstScanner(ScanNullString)
	ScanInts            = NewSliceScanner(ScanAny[int])
	ScanInt32s          = NewSliceScanner(ScanAny[int32])
	ScanInt64s          = NewSliceScanner(ScanAny[int64])
	Scanuint32s         = NewSliceScanner(ScanAny[uint32])
	ScanFirstInt        = NewFirstScanner(ScanAny[int])
	ScanFirstInt64      = NewFirstScanner(ScanAny[int64])
	ScanFirstNullInt64  = NewFirstScanner(ScanNullInt64)
	ScanFloats          = NewSliceScanner(ScanAny[float64])
	ScanFirstFloat      = NewFirstScanner(ScanAny[float64])
	ScanBools           = NewSliceScanner(ScanAny[bool])
	ScanFirstBool       = NewFirstScanner(ScanAny[bool])
	ScanTimes           = NewSliceScanner(ScanAny[time.Time])
	ScanFirstTime       = NewFirstScanner(ScanAny[time.Time])
)
