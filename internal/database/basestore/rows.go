package basestore

import (
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type Rows interface {
	Next() bool
	Close() error
	Err() error
	Scan(...interface{}) error
}

// CloseRows closes the given rows object. The resulting error is a multierror
// containing the error parameter along with any errors that occur during scanning
// or closing the rows object. The rows object is assumed to be non-nil.
//
// The signature of this function allows scan methods to be written uniformly:
//
//	func ScanThings(rows *sql.Rows, queryErr error) (_ []Thing, err error) {
//	    if queryErr != nil {
//	        return nil, queryErr
//	    }
//	    defer func() { err = CloseRows(rows, err) }()
//
//	    // read things from rows
//	}
//
// Scan methods should be called directly with the results of `*store.Query` to
// ensure that the rows are always properly handled.
//
//	things, err := ScanThings(store.Query(ctx, query))
func CloseRows(rows Rows, err error) error {
	return errors.Append(err, rows.Close(), rows.Err())
}
