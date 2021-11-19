package symbols

import "github.com/jmoiron/sqlx"

// SanityCheck makes sure that go-sqlite3 was compiled with cgo by
// seeing if we can actually create a table.
func SanityCheck() error {
	db, err := sqlx.Open("sqlite3_with_regexp", ":memory:")
	if err != nil {
		return err
	}
	defer db.Close()
	_, err = db.Exec("CREATE TABLE test (col TEXT);")
	if err != nil {
		// If go-sqlite3 was not compiled with cgo, the error will be:
		//
		// > Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work. This is a stub
		return err
	}

	return nil
}
