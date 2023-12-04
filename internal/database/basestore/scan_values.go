package basestore

import (
	"database/sql"
	"time"

	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

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

// ScanInt32Array scans a single int32 array from the given scanner.
func ScanInt32Array(s dbutil.Scanner) ([]int32, error) {
	var value pq.Int32Array
	if err := s.Scan(&value); err != nil {
		return nil, err
	}

	return value, nil
}

var (
	ScanInt             = ScanAny[int]
	ScanStrings         = NewSliceScanner(ScanAny[string])
	ScanFirstString     = NewFirstScanner(ScanAny[string])
	ScanNullStrings     = NewSliceScanner(ScanNullString)
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
	ScanNullTimes       = NewSliceScanner(ScanAny[*time.Time])
	ScanFirstNullTime   = NewFirstScanner(ScanAny[*time.Time])
	ScanFirstInt32Array = NewFirstScanner(ScanInt32Array)
)
