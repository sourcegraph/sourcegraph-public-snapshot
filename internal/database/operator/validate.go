package operator

import (
	"database/sql"
	"fmt"
	"net/url"

	"github.com/Masterminds/semver"

	"github.com/sourcegraph/sourcegraph/internal/database/dbconn"
)

// Create a readonly connection to the databases and check that they're schemas are in the expected state.
// Validate that the expected definitions have been defined
func Validate(version *semver.Version, dsn *url.URL) error {
	db, err := dbconn.ConnectInternal(nil, dsn.String(), "operator", "pgsql")
	if err != nil {
		return err
	}

	_, err = db.Exec("SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY;")
	if err != nil {
		return err
	}

	_, err = db.Exec("CREATE TABLE op_readonly_check (id int);")
	if err == nil {
		db.Exec("DROP TABLE op_readonly_check;")
		return fmt.Errorf("Validate requires the database to be in read-only mode for safety")
	}

	fmt.Println("connected to db")
	err = db.Ping()
	if err != nil {
		return err
	}

	err = checkVersion(db, version)
	if err != nil {
		return err
	}

	return nil
}

func checkVersion(db *sql.DB, version *semver.Version) error {
	if version != nil {
		var dbVersion string
		row, err := db.Query(`SELECT version FROM versions;`)
		if err != nil {
			return err
		}
		if !row.Next() {
			return fmt.Errorf("no version found")
		}
		if err = row.Scan(&dbVersion); err != nil {
			return err
		}
		fmt.Println("version", dbVersion)

		if dbVersion != version.String() {
			return fmt.Errorf("version mismatch: %s != %s", dbVersion, version.String())
		}
		fmt.Println("database version,", dbVersion, "is correct")
	}
	return nil
}
