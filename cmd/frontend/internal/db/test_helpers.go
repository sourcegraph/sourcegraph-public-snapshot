package db

import (
	"log"
	"os/exec"
	"strings"
)

// InitTest creates a test database, named with the given suffix, if one does not already exist and
// configures this package to use it. It is called by integration tests (in a package init func)
// that need to use a real database.
func InitTest(nameSuffix string) {
	dbname := "sourcegraph-test-" + nameSuffix

	out, err := exec.Command("createdb", dbname).CombinedOutput()
	if err != nil {
		if strings.Contains(string(out), "already exists") {
			log.Printf("DB %s exists already (run `dropdb %s` to delete and force re-creation)", dbname, dbname)
		} else {
			log.Fatalf("createdb failed: %v\n%s", err, string(out))
		}
	}

	ConnectToDB("dbname=" + dbname)
}
